package logger

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/umutto/birdwatcher/shared/env"
)

// Config holds configuration parameters for initializing a slog.Logger.
// If Output is nil, it defaults to os.Stdout.
type Config struct {
	Env    env.Environment
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
