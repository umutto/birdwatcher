// Package logger provides structured logging initialization and context propagation for Birdwatcher services.
package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"

	"github.com/umutto/birdwatcher/shared/env"
)

const ansiBrightRed = 9

// New initializes and returns a structured slog.Logger based on the provided Config.
func New(cfg Config) *slog.Logger {
	var handler slog.Handler

	w := cfg.Output
	if w == nil {
		w = os.Stdout
	}

	if cfg.Env == env.ProdEnv {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: cfg.Level,
		})
	} else {
		wd, _ := os.Getwd()
		handler = tint.NewTextHandler(w, &tint.Options{
			Level:       cfg.Level,
			TimeFormat:  time.TimeOnly,
			AddSource:   true,
			NoColor:     !shouldUseColor(w),
			ReplaceAttr: replaceDevAttr(wd),
		})
	}

	return slog.New(&contextHandler{Handler: handler})
}

// shouldUseColor reports whether ANSI colors should be enabled for w.
func shouldUseColor(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// replaceDevAttr shortens source paths relative to wd and highlights error attributes in red.
func replaceDevAttr(wd string) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if len(groups) == 0 && a.Key == slog.SourceKey {
			if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
				return slog.Any(a.Key, &slog.Source{
					Function: src.Function,
					File:     shortenPath(wd, src.File),
					Line:     src.Line,
				})
			}
			return a
		}

		if a.Value.Kind() == slog.KindAny {
			if _, ok := a.Value.Any().(error); ok {
				return tint.Attr(ansiBrightRed, a)
			}
		}

		return a
	}
}

// shortenPath returns file relative to wd when nearby, falling back to the base filename.
func shortenPath(wd, file string) string {
	if wd == "" {
		return filepath.Base(file)
	}
	rel, err := filepath.Rel(wd, file)
	if err != nil || strings.HasPrefix(rel, "../../../") {
		return filepath.Base(file)
	}
	return rel
}
