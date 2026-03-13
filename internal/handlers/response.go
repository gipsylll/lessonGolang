package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"sushkov/internal/logger"

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
