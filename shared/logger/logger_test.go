package logger_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/umutto/birdwatcher/shared/env"
	"github.com/umutto/birdwatcher/shared/logger"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLvl slog.Level
		wantErr bool
	}{
		{name: "debug lowercase", input: "debug", wantLvl: slog.LevelDebug, wantErr: false},
		{name: "info lowercase", input: "info", wantLvl: slog.LevelInfo, wantErr: false},
		{name: "warn uppercase", input: "WARN", wantLvl: slog.LevelWarn, wantErr: false},
		{name: "empty string defaults to info", input: "", wantLvl: slog.LevelInfo, wantErr: false},
		{name: "invalid string", input: "NOT_A_LEVEL", wantLvl: slog.LevelInfo, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lvl, err := logger.ParseLevel(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if lvl != tt.wantLvl {
				t.Errorf("ParseLevel(%q) got %v, want %v", tt.input, lvl, tt.wantLvl)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Run("environment format", func(t *testing.T) {
		tests := []struct {
			name         string
			env          env.Environment
			wantContains string
		}{
			{name: "dev text format", env: env.DevEnv, wantContains: `msg="hello world"`},
			{name: "prod json format", env: env.ProdEnv, wantContains: `"msg":"hello world"`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				out := &bytes.Buffer{}
				log := logger.New(logger.Config{Env: tt.env, Output: out})

				log.Info("hello world")
				got := out.String()

				if !strings.Contains(got, tt.wantContains) {
					t.Errorf("got %q, want to contain %q", got, tt.wantContains)
				}
			})
		}
	})

	t.Run("level filtering", func(t *testing.T) {
		out := &bytes.Buffer{}
		log := logger.New(logger.Config{Env: env.DevEnv, Level: slog.LevelWarn, Output: out})

		log.Info("should be ignored")

		if got := out.String(); got != "" {
			t.Errorf("got %q, want empty output for filtered log", got)
		}
	})
}

func TestContextHandler(t *testing.T) {
	t.Run("with job_id", func(t *testing.T) {
		out := &bytes.Buffer{}
		log := logger.New(logger.Config{Env: env.DevEnv, Output: out})
		ctx := logger.WithJobID(context.Background(), "test-job-123")

		log.InfoContext(ctx, "hello world")
		got := out.String()
		want := "job_id=test-job-123"

		if !strings.Contains(got, want) {
			t.Errorf("got %q, want to contain %q", got, want)
		}
	})

	t.Run("without job_id", func(t *testing.T) {
		out := &bytes.Buffer{}
		log := logger.New(logger.Config{Env: env.DevEnv, Output: out})

		log.InfoContext(context.Background(), "hello world")
		got := out.String()

		if strings.Contains(got, "job_id") {
			t.Errorf("got %q, did not want job_id", got)
		}
	})

	t.Run("preserves context with child logger", func(t *testing.T) {
		out := &bytes.Buffer{}
		log := logger.New(logger.Config{Env: env.DevEnv, Output: out})
		child := log.With(slog.String("component", "crawler"))

		ctx := logger.WithJobID(context.Background(), "test-job-123")
		child.InfoContext(ctx, "hello world")
		got := out.String()

		if !strings.Contains(got, "job_id=test-job-123") || !strings.Contains(got, "component=crawler") {
			t.Errorf("got %q, want both job_id and component", got)
		}
	})
}
