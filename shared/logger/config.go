package logger

import (
	"fmt"
	"io"
	"log/slog"
)

// Env represents the operational environment (e.g. development or production).
type Env string

const (
	// DevEnv indicates a local development environment with human-readable text logs.
	DevEnv Env = "development"
	// ProdEnv indicates a production environment with structured JSON logs.
	ProdEnv Env = "production"
)

// Config holds configuration parameters for initializing a slog.Logger.
// If Output is nil, it defaults to os.Stdout.
type Config struct {
	Env    Env
	Level  slog.Level
	Output io.Writer
}

// ParseLevel converts a level string into a slog.Level, defaulting to Info if empty or invalid.
func ParseLevel(s string) (slog.Level, error) {
	var lvl slog.Level

	if s == "" {
		return slog.LevelInfo, nil
	}

	err := lvl.UnmarshalText([]byte(s))

	if err != nil {
		return slog.LevelInfo, fmt.Errorf("invalid log level %q: %w", s, err)
	}

	return lvl, nil
}
