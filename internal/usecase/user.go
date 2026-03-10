package usecase

import (
	"context"
	"strconv"
	"strings"

	"sushkov/internal/domain"
)

type UserRepository interface {
	GetAll(ctx context.Context) ([]domain.User, error)
	List(ctx context.Context, input domain.ListUsersInput) (domain.UserPage, error)
	GetByID(ctx context.Context, id int) (domain.User, error)
	Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error)
	Update(ctx context.Context, id, version int, input domain.UpdateUserInput) (domain.User, error)
	Patch(ctx context.Context, id, version int, input domain.PatchUserInput) (domain.User, error)
}

type UserUsecase struct {
	repo UserRepository
}

func NewUserUsecase(repo UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) GetAll(ctx context.Context) ([]domain.User, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *UserUsecase) List(ctx context.Context, input domain.ListUsersInput) (domain.UserPage, error) {
	return uc.repo.List(ctx, input)
}

func (uc *UserUsecase) GetByID(ctx context.Context, id int) (domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) Create(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	return uc.repo.Create(ctx, input)
}

func (uc *UserUsecase) Update(ctx context.Context, id int, ifMatch string, input domain.UpdateUserInput) (domain.User, error) {
	version, err := parseIfMatch(ifMatch)
	if err != nil {
		return domain.User{}, err
	}
	return uc.repo.Update(ctx, id, version, input)
}

func (uc *UserUsecase) Patch(ctx context.Context, id int, ifMatch string, input domain.PatchUserInput) (domain.User, error) {
	version, err := parseIfMatch(ifMatch)
	if err != nil {
		return domain.User{}, err
	}
	return uc.repo.Patch(ctx, id, version, input)
}

// parseIfMatch разбирает заголовок If-Match вида `"v2"` и возвращает номер версии.
func parseIfMatch(ifMatch string) (int, error) {
	if ifMatch == "" {
		return 0, domain.ErrPreconditionRequired
	}
	s := strings.Trim(ifMatch, `"`)
	s = strings.TrimPrefix(s, "v")
	version, err := strconv.Atoi(s)
	if err != nil {
		return 0, domain.ErrInvalidETag
	}
	return version, nil
}
