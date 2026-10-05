package env_test

import (
	"errors"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/shared/env"
)

func TestCurrent(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want env.Environment
	}{
		{name: "explicit development environment", val: "development", want: env.DevEnv},
		{name: "test environment", val: "test", want: env.TestEnv},
		{name: "production environment", val: "production", want: env.ProdEnv},
		{name: "case-insensitive and trimmed whitespace", val: "  PRODUCTION  ", want: env.ProdEnv},
		{name: "empty environment value", val: "", want: env.DevEnv},
		{name: "unknown environment value", val: "unknown", want: env.DevEnv},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENV", tt.val)

			got := env.Current()
			if got != tt.want {
				t.Errorf("Current() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequire(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		const key = "TEST_REQUIRE_KEY"
		want := "hello"
		t.Setenv(key, want)

		got, err := env.Require(key)
		if err != nil {
			t.Fatalf("Require(%q) unexpected error: %v", key, err)
		}
		if got != want {
			t.Errorf("Require(%q) = %q, want %q", key, got, want)
		}
	})

	t.Run("returns ErrMissing when unset", func(t *testing.T) {
		const key = "UNSET_REQUIRE_KEY"

		_, err := env.Require(key)
		if !errors.Is(err, env.ErrMissing) {
			t.Errorf("Require(%q) error = %v, want %v", key, err, env.ErrMissing)
		}
	})

	t.Run("returns ErrMissing when empty string", func(t *testing.T) {
		const key = "EMPTY_REQUIRE_KEY"
		t.Setenv(key, "")

		_, err := env.Require(key)
		if !errors.Is(err, env.ErrMissing) {
			t.Errorf("Require(%q) error = %v, want %v", key, err, env.ErrMissing)
		}
	})
}

func TestString(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		const key = "STRING_KEY"
		want := "hello"
		fallback := "fallback"
		t.Setenv(key, want)

		got := env.String(key, fallback)
		if got != want {
			t.Errorf("String(%q, %q) = %q, want %q", key, fallback, got, want)
		}
	})

	t.Run("uses fallback when unset", func(t *testing.T) {
		const key = "UNSET_STRING_KEY"
		fallback := "fallback"

		got := env.String(key, fallback)
		if got != fallback {
			t.Errorf("String(%q, %q) = %q, want %q", key, fallback, got, fallback)
		}
	})

	t.Run("uses fallback when empty", func(t *testing.T) {
		const key = "EMPTY_STRING_KEY"
		fallback := "fallback"
		t.Setenv(key, "")

		got := env.String(key, fallback)
		if got != fallback {
			t.Errorf("String(%q, %q) = %q, want %q", key, fallback, got, fallback)
		}
	})
}

func TestInt32(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		const key = "INT_KEY"
		want := int32(42)
		fallback := int32(10)
		t.Setenv(key, "42")

		got, err := env.Int32(key, fallback)
		if err != nil {
			t.Fatalf("Int32(%q, %d) unexpected error: %v", key, fallback, err)
		}
		if got != want {
			t.Errorf("Int32(%q, %d) = %d, want %d", key, fallback, got, want)
		}
	})

	t.Run("uses fallback when unset", func(t *testing.T) {
		const key = "UNSET_INT_KEY"
		fallback := int32(10)

		got, err := env.Int32(key, fallback)
		if err != nil {
			t.Fatalf("Int32(%q, %d) unexpected error: %v", key, fallback, err)
		}
		if got != fallback {
			t.Errorf("Int32(%q, %d) = %d, want %d", key, fallback, got, fallback)
		}
	})

	t.Run("uses fallback when empty", func(t *testing.T) {
		const key = "EMPTY_INT_KEY"
		fallback := int32(10)
		t.Setenv(key, "")

		got, err := env.Int32(key, fallback)
		if err != nil {
			t.Fatalf("Int32(%q, %d) unexpected error: %v", key, fallback, err)
		}
		if got != fallback {
			t.Errorf("Int32(%q, %d) = %d, want %d", key, fallback, got, fallback)
		}
	})

	t.Run("returns error when invalid", func(t *testing.T) {
		const key = "INVALID_INT_KEY"
		fallback := int32(10)
		t.Setenv(key, "banana")

		got, err := env.Int32(key, fallback)
		if err == nil {
			t.Fatalf("Int32(%q, %d) expected error, got nil", key, fallback)
		}
		if got != fallback {
			t.Errorf("Int32(%q, %d) = %d, want fallback %d", key, fallback, got, fallback)
		}
	})
}

func TestDuration(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		const key = "DURATION_KEY"
		want := 10 * time.Second
		fallback := 5 * time.Second
		t.Setenv(key, "10s")

		got, err := env.Duration(key, fallback)
		if err != nil {
			t.Fatalf("Duration(%q, %v) unexpected error: %v", key, fallback, err)
		}
		if got != want {
			t.Errorf("Duration(%q, %v) = %v, want %v", key, fallback, got, want)
		}
	})

	t.Run("uses fallback when unset", func(t *testing.T) {
		const key = "UNSET_DURATION_KEY"
		fallback := 5 * time.Second

		got, err := env.Duration(key, fallback)
		if err != nil {
			t.Fatalf("Duration(%q, %v) unexpected error: %v", key, fallback, err)
		}
		if got != fallback {
			t.Errorf("Duration(%q, %v) = %v, want %v", key, fallback, got, fallback)
		}
	})

	t.Run("uses fallback when empty", func(t *testing.T) {
		const key = "EMPTY_DURATION_KEY"
		fallback := 5 * time.Second
		t.Setenv(key, "")

		got, err := env.Duration(key, fallback)
		if err != nil {
			t.Fatalf("Duration(%q, %v) unexpected error: %v", key, fallback, err)
		}
		if got != fallback {
			t.Errorf("Duration(%q, %v) = %v, want %v", key, fallback, got, fallback)
		}
	})

	t.Run("returns error when invalid", func(t *testing.T) {
		const key = "INVALID_DURATION_KEY"
		fallback := 5 * time.Second
		t.Setenv(key, "not-a-duration")

		got, err := env.Duration(key, fallback)
		if err == nil {
			t.Fatalf("Duration(%q, %v) expected error, got nil", key, fallback)
		}
		if got != fallback {
			t.Errorf("Duration(%q, %v) = %v, want fallback %v", key, fallback, got, fallback)
		}
	})
}

func TestBool(t *testing.T) {
	t.Run("returns value when set", func(t *testing.T) {
		const key = "BOOL_KEY"
		want := true
		fallback := false
		t.Setenv(key, "true")

		got, err := env.Bool(key, fallback)
		if err != nil {
			t.Fatalf("Bool(%q, %t) unexpected error: %v", key, fallback, err)
		}
		if got != want {
			t.Errorf("Bool(%q, %t) = %t, want %t", key, fallback, got, want)
		}
	})

	t.Run("uses fallback when unset", func(t *testing.T) {
		const key = "UNSET_BOOL_KEY"
		fallback := false

		got, err := env.Bool(key, fallback)
		if err != nil {
			t.Fatalf("Bool(%q, %t) unexpected error: %v", key, fallback, err)
		}
		if got != fallback {
			t.Errorf("Bool(%q, %t) = %t, want %t", key, fallback, got, fallback)
		}
	})

	t.Run("uses fallback when empty", func(t *testing.T) {
		const key = "EMPTY_BOOL_KEY"
		fallback := false
		t.Setenv(key, "")

		got, err := env.Bool(key, fallback)
		if err != nil {
			t.Fatalf("Bool(%q, %t) unexpected error: %v", key, fallback, err)
		}
		if got != fallback {
			t.Errorf("Bool(%q, %t) = %t, want %t", key, fallback, got, fallback)
		}
	})

	t.Run("returns error when invalid", func(t *testing.T) {
		const key = "INVALID_BOOL_KEY"
		fallback := false
		t.Setenv(key, "not-a-bool")

		got, err := env.Bool(key, fallback)
		if err == nil {
			t.Fatalf("Bool(%q, %t) expected error, got nil", key, fallback)
		}
		if got != fallback {
			t.Errorf("Bool(%q, %t) = %t, want fallback %t", key, fallback, got, fallback)
		}
	})
}
