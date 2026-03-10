package domain

import "context"

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Version int    `json:"version"`
}

// --- Input types (передаются из handler в usecase) ---

type CreateUserInput struct {
	Name  string
	Email string
}

type UpdateUserInput struct {
	Name  string
	Email string
}

type PatchUserInput struct {
	Name  *string
	Email *string
}

// --- Порты (интерфейсы) ---

// UserRepository — порт для хранилища. Реализуется в adapter/.
type UserRepository interface {
	GetAll(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id int) (User, error)
	Create(ctx context.Context, input CreateUserInput) (User, error)
	Update(ctx context.Context, id, version int, input UpdateUserInput) (User, error)
	Patch(ctx context.Context, id, version int, input PatchUserInput) (User, error)
}

// UserUsecase — порт для бизнес-логики. Реализуется в usecase/.
type UserUsecase interface {
	GetAll(ctx context.Context) ([]User, error)
	GetByID(ctx context.Context, id int) (User, error)
	Create(ctx context.Context, input CreateUserInput) (User, error)
	Update(ctx context.Context, id int, ifMatch string, input UpdateUserInput) (User, error)
	Patch(ctx context.Context, id int, ifMatch string, input PatchUserInput) (User, error)
}
