package infrastructure

import (
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

func RunMigrations(pool *pgxpool.Pool, migrationsFS fs.FS) error {
	db := stdlib.OpenDBFromPool(pool)

	goose.SetLogger(goose.NopLogger())
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, "."); err != nil {
		return err
	}

	log.Info().Msg("database migrations applied")
	return nil
}
