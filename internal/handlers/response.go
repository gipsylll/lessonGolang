package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sushkov/internal/logger"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id"`
	Fields    []fieldError `json:"fields,omitempty"`
}

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	requestID := logger.RequestIDFromContext(r.Context())

	log.Warn().
		Str("request_id", requestID).
		Str("code", code).
		Str("message", message).
		Int("status", status).
		Msg("error response")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encode(w, errorResponse{
		Error: errorBody{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	})
}

func writeValidationError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := logger.RequestIDFromContext(r.Context())

	var fields []fieldError
	for _, e := range err.(validator.ValidationErrors) {
		fields = append(fields, fieldError{
			Field:   strings.ToLower(e.Field()),
			Message: validationMessage(e),
		})
	}

	log.Warn().
		Str("request_id", requestID).
		Int("status", http.StatusUnprocessableEntity).
		Interface("fields", fields).
		Msg("validation error")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	encode(w, errorResponse{
		Error: errorBody{
			Code:      "validation_error",
			Message:   "request body is invalid",
			RequestID: requestID,
			Fields:    fields,
		},
	})
}

func writeFieldErrors(w http.ResponseWriter, r *http.Request, fields []fieldError) {
	requestID := logger.RequestIDFromContext(r.Context())

	log.Warn().
		Str("request_id", requestID).
		Interface("fields", fields).
		Msg("validation error")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	encode(w, errorResponse{
		Error: errorBody{
			Code:      "validation_error",
			Message:   "request body is invalid",
			RequestID: requestID,
			Fields:    fields,
		},
	})
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
	case "oneof":
		return "must be one of: " + e.Param()
	default:
		return "invalid value"
	}
}

// etagFor формирует ETag по версии: "v1", "v2", ...
func etagFor(version int) string {
	return fmt.Sprintf(`"v%d"`, version)
}

func writeOk(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encode(w, data)
}

// encode пишет JSON в ответ и логирует если что-то пошло не так.
// После отправки заголовков мы не можем изменить статус, поэтому только логируем.
func encode(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
	}
}
