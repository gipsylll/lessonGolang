package usecase

import (
	"context"
	"strconv"
	"strings"

	"sushkov/internal/domain"
	"sushkov/internal/interfaces"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

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
	if err := validateStruct(input); err != nil {
		return domain.User{}, err
	}
	return uc.repo.Create(ctx, input)
}

func (uc *UserUsecase) Update(ctx context.Context, id int, ifMatch string, input interfaces.UpdateUserInput) (domain.User, error) {
	if err := validateStruct(input); err != nil {
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
	if err := validateStruct(input); err != nil {
		return domain.User{}, err
	}
	version, err := parseIfMatch(ifMatch)
	if err != nil {
		return domain.User{}, err
	}
	return uc.repo.Patch(ctx, id, version, input)
}

func validateStruct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return &domain.ValidationError{Fields: []domain.FieldError{{Field: "", Message: err.Error()}}}
	}
	fields := make([]domain.FieldError, len(errs))
	for i, e := range errs {
		fields[i] = domain.FieldError{
			Field:   strings.ToLower(e.Field()),
			Message: validationMessage(e),
		}
	}
	return &domain.ValidationError{Fields: fields}
}

func validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "min":
		return "value is too short (min " + e.Param() + ")"
	case "max":
		return "value is too long (max " + e.Param() + ")"
	case "email":
		return "must be a valid email"
	case "omitempty":
		return "invalid value"
	default:
		return "invalid value"
	}
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
