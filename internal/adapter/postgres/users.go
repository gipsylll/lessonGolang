package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"sushkov/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const pgUniqueViolation = "23505"

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *UserRepo) List(ctx context.Context, input domain.ListUsersInput) (domain.UserPage, error) {
	pageSize := input.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 15
	}

	var (
		rows pgx.Rows
		err  error
	)

	if input.Cursor == "" {
		rows, err = r.db.Query(ctx,
			`SELECT id, name, email, version, created_at FROM users
			 ORDER BY created_at, id LIMIT $1`,
			pageSize+1,
		)
	} else {
		cursorTime, cursorID, parseErr := parseCursor(input.Cursor)
		if parseErr != nil {
			return domain.UserPage{}, domain.ErrInvalidCursor
		}
		rows, err = r.db.Query(ctx,
			`SELECT id, name, email, version, created_at FROM users
			 WHERE (created_at, id) > ($1, $2)
			 ORDER BY created_at, id LIMIT $3`,
			cursorTime, cursorID, pageSize+1,
		)
	}
	if err != nil {
		return domain.UserPage{}, err
	}
	defer rows.Close()

	var items []domain.User
	var createdAts []time.Time
	for rows.Next() {
		var u domain.User
		var ca time.Time
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Version, &ca); err != nil {
			return domain.UserPage{}, err
		}
		items = append(items, u)
		createdAts = append(createdAts, ca)
	}
	if err := rows.Err(); err != nil {
		return domain.UserPage{}, err
	}

	var nextCursor string
	if len(items) > pageSize {
		items = items[:pageSize]
		last := items[pageSize-1]
		nextCursor = buildCursor(createdAts[pageSize-1], last.ID)
	}

	return domain.UserPage{Items: items, NextCursor: nextCursor}, nil
}

func parseCursor(cursor string) (time.Time, int, error) {
	parts := strings.SplitN(cursor, "|", 2)
	if len(parts) != 2 {
		return time.Time{}, 0, errors.New("invalid cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, 0, err
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, 0, err
	}
	return t, id, nil
}

func buildCursor(t time.Time, id int) string {
	return t.UTC().Format(time.RFC3339Nano) + "|" + strconv.Itoa(id)
}

func (r *UserRepo) GetAll(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, email, version FROM users ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Version); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) GetByID(ctx context.Context, id int) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		`SELECT id, name, email, version FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Name, &u.Email, &u.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	return u, err
}

func (r *UserRepo) Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	var u domain.User
	err := r.withTx(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx,
			`INSERT INTO users (name, email) VALUES ($1, $2)
			 RETURNING id, name, email, version`,
			input.Name, input.Email,
		).Scan(&u.ID, &u.Name, &u.Email, &u.Version)
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		if err != nil {
			return err
		}
		return insertAuditLog(ctx, tx, u.ID, "create", nil, u)
	})
	return u, err
}

func (r *UserRepo) Update(ctx context.Context, id, version int, input domain.UpdateUserInput) (domain.User, error) {
	var updated domain.User
	err := r.withTx(ctx, func(tx pgx.Tx) error {
		var old domain.User
		err := tx.QueryRow(ctx,
			`SELECT id, name, email, version FROM users WHERE id = $1 FOR UPDATE`,
			id,
		).Scan(&old.ID, &old.Name, &old.Email, &old.Version)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		if old.Version != version {
			return domain.ErrPreconditionFailed
		}

		err = tx.QueryRow(ctx,
			`UPDATE users
			 SET name = $1, email = $2, version = version + 1, updated_at = NOW()
			 WHERE id = $3
			 RETURNING id, name, email, version`,
			input.Name, input.Email, id,
		).Scan(&updated.ID, &updated.Name, &updated.Email, &updated.Version)
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		if err != nil {
			return err
		}

		return insertAuditLog(ctx, tx, id, "update", &old, updated)
	})
	return updated, err
}

func (r *UserRepo) Patch(ctx context.Context, id, version int, input domain.PatchUserInput) (domain.User, error) {
	var updated domain.User
	err := r.withTx(ctx, func(tx pgx.Tx) error {
		var old domain.User
		err := tx.QueryRow(ctx,
			`SELECT id, name, email, version FROM users WHERE id = $1 FOR UPDATE`,
			id,
		).Scan(&old.ID, &old.Name, &old.Email, &old.Version)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		if old.Version != version {
			return domain.ErrPreconditionFailed
		}

		err = tx.QueryRow(ctx,
			`UPDATE users
			 SET name       = COALESCE($1, name),
			     email      = COALESCE($2, email),
			     version    = version + 1,
			     updated_at = NOW()
			 WHERE id = $3
			 RETURNING id, name, email, version`,
			input.Name, input.Email, id,
		).Scan(&updated.ID, &updated.Name, &updated.Email, &updated.Version)
		if isUniqueViolation(err) {
			return domain.ErrEmailTaken
		}
		if err != nil {
			return err
		}

		return insertAuditLog(ctx, tx, id, "patch", &old, updated)
	})
	return updated, err
}

func insertAuditLog(ctx context.Context, tx pgx.Tx, entityID int, action string, old *domain.User, new domain.User) error {
	var oldJSON []byte
	if old != nil {
		var err error
		oldJSON, err = json.Marshal(old)
		if err != nil {
			return err
		}
	}
	newJSON, err := json.Marshal(new)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO audit_log (entity_id, action, old_data, new_data) VALUES ($1, $2, $3, $4)`,
		entityID, action, oldJSON, newJSON,
	)
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}
