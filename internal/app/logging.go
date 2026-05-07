package app

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func (w *responseWriter) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

// LoggingMiddleware logs method, URI, duration, status, and response size with the provided logger.
func LoggingMiddleware(log zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, status: 0}
			next.ServeHTTP(wrapped, r)
			duration := time.Since(start)
			log.Info().
				Str("uri", r.URL.RequestURI()).
				Str("method", r.Method).
				Dur("duration", duration).
				Int("status", wrapped.statusCode()).
				Int("size", wrapped.size).
				Msg("request")
		})
	}
}
