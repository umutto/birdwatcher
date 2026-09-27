// Package logger provides structured logging initialization and context propagation for Birdwatcher services.
package logger

import (
	"log/slog"
	"os"
)

// New initializes and returns a structured slog.Logger based on the provided Config.
func New(cfg Config) *slog.Logger {
	var handler slog.Handler

	w := cfg.Output
	if w == nil {
		w = os.Stdout
	}

	if cfg.Env == ProdEnv {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	} else {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			Level:     cfg.Level,
			AddSource: true,
		})
	}

	return slog.New(&contextHandler{Handler: handler})
}
