package server

import (
	"log/slog"
	"net/http"
	"time"
)

// Middleware represents an HTTP middleware function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// responseWriter wraps http.ResponseWriter to capture the HTTP status code and response size.
type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

// newResponseWriter initializes and returns a responseWriter wrapping w.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader captures the HTTP status code and forwards it to the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write writes data to the connection and accumulates the number of bytes written.
func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// requestLogger returns an HTTP middleware that logs incoming requests, status codes, and latencies.
func requestLogger(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			level := slog.LevelInfo
			switch {
			case rw.status >= 500:
				level = slog.LevelError
			case rw.status >= 400:
				level = slog.LevelWarn
			}

			log.LogAttrs(r.Context(), level, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Duration("duration", time.Since(start)),
				slog.Int("bytes", rw.bytes),
			)
		})
	}
}

// chain wraps an http.Handler with a sequence of Middlewares executed in slice order.
func chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
