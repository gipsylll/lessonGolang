package usecase

import (
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"

	"sushkov/internal/domain"
	"sushkov/internal/interfaces"
)

type UserUsecase struct {
	repo interfaces.UserRepository
}

func NewUserUsecase(repo interfaces.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (uc *UserUsecase) GetAll(ctx context.Context) ([]domain.User, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *UserUsecase) List(ctx context.Context, input interfaces.ListUsersInput) (interfaces.UserPage, error) {
	return uc.repo.List(ctx, input)
}

func (uc *UserUsecase) GetByID(ctx context.Context, id int) (domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserUsecase) Create(ctx context.Context, input interfaces.CreateUserInput) (domain.User, error) {
	if err := validateUserFields(input.Name, input.Email); err != nil {
		return domain.User{}, err
	}
	return uc.repo.Create(ctx, input)
}

func (uc *UserUsecase) Update(ctx context.Context, id int, ifMatch string, input interfaces.UpdateUserInput) (domain.User, error) {
	if err := validateUserFields(input.Name, input.Email); err != nil {
		return domain.User{}, err
	}
	version, err := parseIfMatch(ifMatch)
	if err != nil {
		return domain.User{}, err
	}
	return uc.repo.Update(ctx, id, version, input)
}

func (uc *UserUsecase) Patch(ctx context.Context, id int, ifMatch string, input interfaces.PatchUserInput) (domain.User, error) {
	if input.Name == nil && input.Email == nil {
		return domain.User{}, domain.ErrNoFieldsToUpdate
	}
	var fieldErrs []domain.FieldError
	if input.Name != nil {
		if err := validateName(*input.Name); err != nil {
			fieldErrs = append(fieldErrs, domain.FieldError{Field: "name", Message: err.Error()})
		}
	}
	if input.Email != nil {
		if err := validateEmail(*input.Email); err != nil {
			fieldErrs = append(fieldErrs, domain.FieldError{Field: "email", Message: err.Error()})
		}
	}
	if len(fieldErrs) > 0 {
		return domain.User{}, &domain.ValidationError{Fields: fieldErrs}
	}
	version, err := parseIfMatch(ifMatch)
	if err != nil {
		return domain.User{}, err
	}
	return uc.repo.Patch(ctx, id, version, input)
}

func validateUserFields(name, email string) error {
	var fieldErrs []domain.FieldError
	if err := validateName(name); err != nil {
		fieldErrs = append(fieldErrs, domain.FieldError{Field: "name", Message: err.Error()})
	}
	if err := validateEmail(email); err != nil {
		fieldErrs = append(fieldErrs, domain.FieldError{Field: "email", Message: err.Error()})
	}
	if len(fieldErrs) > 0 {
		return &domain.ValidationError{Fields: fieldErrs}
	}
	return nil
}

func validateName(name string) error {
	l := len(name)
	if l < 2 {
		return errors.New("must be at least 2 characters")
	}
	if l > 100 {
		return errors.New("must be at most 100 characters")
	}
	return nil
}

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return errors.New("must be a valid email address")
	}
	return nil
}

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
