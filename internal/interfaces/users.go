package interfaces

import (
	"context"

	"sushkov/internal/domain"
)

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

type ListUsersInput struct {
	PageSize int
	Cursor   string
}

type UserPage struct {
	Items      []domain.User `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

type UserRepository interface {
	GetAll(ctx context.Context) ([]domain.User, error)
	List(ctx context.Context, input ListUsersInput) (UserPage, error)
	GetByID(ctx context.Context, id int) (domain.User, error)
	Create(ctx context.Context, input CreateUserInput) (domain.User, error)
	Update(ctx context.Context, id, version int, input UpdateUserInput) (domain.User, error)
	Patch(ctx context.Context, id, version int, input PatchUserInput) (domain.User, error)
}

type UserUsecase interface {
	List(ctx context.Context, input ListUsersInput) (UserPage, error)
	GetByID(ctx context.Context, id int) (domain.User, error)
	Create(ctx context.Context, input CreateUserInput) (domain.User, error)
	Update(ctx context.Context, id int, ifMatch string, input UpdateUserInput) (domain.User, error)
	Patch(ctx context.Context, id int, ifMatch string, input PatchUserInput) (domain.User, error)
}
