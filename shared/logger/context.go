package logger

import (
	"context"
	"log/slog"
)

// contextKey defines a private context key type to prevent context collisions.
type contextKey struct{}

var jobIDKey = contextKey{}

// WithJobID stores a jobID into the context.
func WithJobID(ctx context.Context, jobID string) context.Context {
	return context.WithValue(ctx, jobIDKey, jobID)
}

// JobIDFromContext retrieves the jobID from the context.
func JobIDFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(jobIDKey).(string)
	return val, ok
}

// contextHandler wraps an slog.Handler to attach context-scoped attributes to log records.
type contextHandler struct {
	slog.Handler
}

// Handle injects the context job ID into the log record before delegating to the wrapped handler.
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if jobID, ok := JobIDFromContext(ctx); ok {
		r.AddAttrs(slog.String("job_id", jobID))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new contextHandler with the given attributes added to the wrapped handler.
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new contextHandler with the given group name added to the wrapped handler.
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}
