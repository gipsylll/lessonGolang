package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"sushkov/internal/logger"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var bufPool = sync.Pool{
	New: func() any {
		b := new(bytes.Buffer)
		b.Grow(512)
		return b
	},
}

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

	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		writeError(w, r, http.StatusUnprocessableEntity, "validation_error", "request body is invalid")
		return
	}

	var fields []fieldError
	for _, e := range errs {
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

func etagFor(version int) string {
	return fmt.Sprintf(`"v%d"`, version)
}

func writeOk(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encode(w, data)
}

func encode(w http.ResponseWriter, v any) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	if err := json.NewEncoder(buf).Encode(v); err != nil {
		log.Error().Err(err).Msg("failed to encode response")
		return
	}
	_, _ = w.Write(buf.Bytes())
}
