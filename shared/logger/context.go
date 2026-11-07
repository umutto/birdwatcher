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

// contextKey defines a private context key type to prevent context collisions.
type contextHandler struct {
	slog.Handler
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if jobID, ok := JobIDFromContext(ctx); ok {
		r.AddAttrs(slog.String("job_id", jobID))
	}
	return h.Handler.Handle(ctx, r)
}
func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}
func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}
