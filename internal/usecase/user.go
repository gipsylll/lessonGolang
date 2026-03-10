package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sushkov/internal/domain"
)

// UserUsecase — реализация domain.UserUsecase.
type UserUsecase struct {
	repo domain.UserRepository
}

func NewUserUsecase(repo domain.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) GetAll(ctx context.Context) ([]domain.User, error) {
	return uc.repo.GetAll(ctx)
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
		return 0, fmt.Errorf("%w: invalid If-Match format", domain.ErrPreconditionFailed)
	}
	return version, nil
}
