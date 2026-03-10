package memory

import (
	"context"
	"sync"

	"sushkov/internal/domain"
)

// UserRepo — in-memory реализация domain.UserRepository.
// Используется до подключения реальной БД.
type UserRepo struct {
	mu    sync.RWMutex
	users []domain.User
}

func NewUserRepo(initial []domain.User) *UserRepo {
	users := make([]domain.User, len(initial))
	copy(users, initial)
	for i := range users {
		users[i].Version = 1
	}
	return &UserRepo{users: users}
}

func (r *UserRepo) GetAll(_ context.Context) ([]domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make([]domain.User, len(r.users))
	copy(snapshot, r.users)
	return snapshot, nil
}

func (r *UserRepo) GetByID(_ context.Context, id int) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if id < 1 || id > len(r.users) {
		return domain.User{}, domain.ErrNotFound
	}
	return r.users[id-1], nil
}

func (r *UserRepo) Create(_ context.Context, input domain.CreateUserInput) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user := domain.User{
		ID:      len(r.users) + 1,
		Name:    input.Name,
		Email:   input.Email,
		Version: 1,
	}
	r.users = append(r.users, user)
	return user, nil
}

func (r *UserRepo) Update(_ context.Context, id, version int, input domain.UpdateUserInput) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id < 1 || id > len(r.users) {
		return domain.User{}, domain.ErrNotFound
	}
	if r.users[id-1].Version != version {
		return domain.User{}, domain.ErrPreconditionFailed
	}

	r.users[id-1].Name = input.Name
	r.users[id-1].Email = input.Email
	r.users[id-1].Version++
	return r.users[id-1], nil
}

func (r *UserRepo) Patch(_ context.Context, id, version int, input domain.PatchUserInput) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if id < 1 || id > len(r.users) {
		return domain.User{}, domain.ErrNotFound
	}
	if r.users[id-1].Version != version {
		return domain.User{}, domain.ErrPreconditionFailed
	}

	if input.Name != nil {
		r.users[id-1].Name = *input.Name
	}
	if input.Email != nil {
		r.users[id-1].Email = *input.Email
	}
	r.users[id-1].Version++
	return r.users[id-1], nil
}
