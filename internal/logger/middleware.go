package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// contextKey — приватный тип для ключей контекста, исключает коллизии с другими пакетами.
type contextKey string

const requestIDKey contextKey = "request_id"

// RequestIDFromContext возвращает request_id из контекста запроса.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// ResponseWriter — обёртка для захвата статуса ответа и размера тела.
type ResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// LoggingMiddleware — middleware для логирования HTTP-запросов.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		wrapped := &ResponseWriter{ResponseWriter: w}

		logger := log.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("ip", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Logger()

		logger.Debug().Msg("request started")

		next.ServeHTTP(wrapped, r)

		logger.Info().
			Int("status", wrapped.statusCode).
			Int("bytes", wrapped.bytesWritten).
			Dur("duration", time.Since(start)).
			Msg("request completed")
	})
}
