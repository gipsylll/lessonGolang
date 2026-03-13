package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"sushkov/internal/domain"
	"sushkov/internal/interfaces"

	"github.com/rs/zerolog/log"
)

type createUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type updateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type patchUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type UserHandler struct {
	uc interfaces.UserUsecase
}

func NewUserHandler(uc interfaces.UserUsecase) *UserHandler {
	return &UserHandler{uc: uc}
}

func (h *UserHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /users", h.GetUsers)
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users/{id}", h.GetUserByID)
	mux.HandleFunc("PUT /users/{id}", h.UpdateUser)
	mux.HandleFunc("PATCH /users/{id}", h.PatchUser)
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "GetUsers").Logger()

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	cursor := r.URL.Query().Get("cursor")

	page, err := h.uc.List(r.Context(), interfaces.ListUsersInput{
		PageSize: pageSize,
		Cursor:   cursor,
	})
	if err != nil {
		writeUCError(w, r, err)
		return
	}

	logger.Debug().Int("count", len(page.Items)).Msg("users list fetched")
	writeOk(w, http.StatusOK, page)
}

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

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context()).With().Str("handler", "CreateUser").Logger()

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	user, err := h.uc.Create(r.Context(), interfaces.CreateUserInput{
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		writeUCError(w, r, err)
		return
	}

	logger.Info().Int("user_id", user.ID).Msg("user created")
	w.Header().Set("ETag", etagFor(user.Version))
	writeOk(w, http.StatusCreated, user)
}

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

	user, err := h.uc.Update(r.Context(), id, r.Header.Get("If-Match"), interfaces.UpdateUserInput{
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

	user, err := h.uc.Patch(r.Context(), id, r.Header.Get("If-Match"), interfaces.PatchUserInput{
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

func pathID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func writeUCError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		fields := make([]fieldError, len(validationErr.Fields))
		for i, f := range validationErr.Fields {
			fields[i] = fieldError{Field: f.Field, Message: f.Message}
		}
		writeFieldErrors(w, r, fields)
		return
	}
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, r, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		writeError(w, r, http.StatusConflict, "email_taken", err.Error())
	case errors.Is(err, domain.ErrPreconditionRequired):
		writeError(w, r, http.StatusPreconditionRequired, "precondition_required", "If-Match header is required; fetch the resource first to get its ETag")
	case errors.Is(err, domain.ErrPreconditionFailed):
		writeError(w, r, http.StatusPreconditionFailed, "precondition_failed", err.Error())
	case errors.Is(err, domain.ErrInvalidETag):
		writeError(w, r, http.StatusBadRequest, "invalid_etag", err.Error())
	case errors.Is(err, domain.ErrInvalidCursor):
		writeError(w, r, http.StatusBadRequest, "invalid_cursor", err.Error())
	case errors.Is(err, domain.ErrNoFieldsToUpdate):
		writeError(w, r, http.StatusBadRequest, "no_fields", err.Error())
	default:
		writeError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error occurred")
	}
}
