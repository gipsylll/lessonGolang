package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"sushkov/internal/domain"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var validate = validator.New()

type createUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

type updateUserRequest struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

// patchUserRequest — поля-указатели, чтобы отличить "не передан" от "передан пустым"
type patchUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type UserHandler struct {
	uc domain.UserUsecase
}

func NewUserHandler(uc domain.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

// Register регистрирует все роуты хендлера в mux.
func (h *UserHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.GetUsers)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users/{id}", h.GetUserByID)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("PATCH /users/{id}", h.PatchUser)
}

// GET /users
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "GetUsers").Logger()

	users, err := h.uc.GetAll(r.Context())
	if err != nil {
		logger.Error().Err(err).Msg("get all users failed")
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to fetch users")
		return
	}

	logger.Debug().Int("count", len(users)).Msg("fetching users list")
	writeOk(w, http.StatusOK, users)
}

// GET /users/{id}
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "GetUserByID").Logger()

	id, err := pathID(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "id must be a positive integer")
		return
	}

	user, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		writeUCError(w, r, err)
		return
	}

	logger.Info().Int("user_id", id).Msg("user found")
	w.Header().Set("ETag", etagFor(user.Version))
	writeOk(w, http.StatusOK, user)
}

// POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "CreateUser").Logger()

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeValidationError(w, r, err)
		return
	}

	user, err := h.uc.Create(r.Context(), domain.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		logger.Error().Err(err).Msg("create user failed")
		writeError(w, r, http.StatusInternalServerError, "internal_error", "failed to create user")
		return
	}

	logger.Info().Int("user_id", user.ID).Msg("user created")
	w.Header().Set("ETag", etagFor(user.Version))
	writeOk(w, http.StatusCreated, user)
}

// PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "UpdateUser").Logger()

	id, err := pathID(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "id must be a positive integer")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeValidationError(w, r, err)
		return
	}

	user, err := h.uc.Update(r.Context(), id, r.Header.Get("If-Match"), domain.UpdateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeUCError(w, r, err)
		return
	}

	logger.Info().Int("user_id", id).Int("version", user.Version).Msg("user updated")
	w.Header().Set("ETag", etagFor(user.Version))
	writeOk(w, http.StatusOK, user)
}

// PATCH /users/{id} — частичное обновление.
// Использует ручные валидационные хелперы (validateLen, validateEmail) из validate.go.
func (h *UserHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "PatchUser").Logger()

	id, err := pathID(r)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_id", "id must be a positive integer")
		return
	}

	var req patchUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	if req.Name == nil && req.Email == nil {
		writeError(w, r, http.StatusBadRequest, "no_fields", "at least one field must be provided")
		return
	}

	var fieldErrs []fieldError
	if req.Name != nil {
		if err := validateLen("name", *req.Name, 2, 100); err != nil {
			fieldErrs = append(fieldErrs, fieldError{Field: "name", Message: err.Error()})
		}
	}
	if req.Email != nil {
		if err := validateEmail(*req.Email); err != nil {
			fieldErrs = append(fieldErrs, fieldError{Field: "email", Message: err.Error()})
		}
	}
	if len(fieldErrs) > 0 {
		writeFieldErrors(w, r, fieldErrs)
		return
	}

	user, err := h.uc.Patch(r.Context(), id, r.Header.Get("If-Match"), domain.PatchUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeUCError(w, r, err)
		return
	}

	logger.Info().Int("user_id", id).Int("version", user.Version).Msg("user patched")
	w.Header().Set("ETag", etagFor(user.Version))
	writeOk(w, http.StatusOK, user)
}

// pathID извлекает {id} из пути и проверяет что он положительный.
func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

// writeUCError транслирует ошибки usecase в HTTP ответы.
func writeUCError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrPreconditionRequired):
		writeError(w, r, http.StatusPreconditionRequired, "precondition_required", err.Error())
	case errors.Is(err, domain.ErrPreconditionFailed):
		writeError(w, r, http.StatusPreconditionFailed, "precondition_failed", err.Error())
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error occurred")
	}
}
