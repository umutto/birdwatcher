// Package env provides utilities for reading and parsing application environment variables.
package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment represents the runtime operational mode (e.g. development, production).
type Environment string

const (
	// DevEnv indicates a local development environment.
	DevEnv Environment = "development"
	// ProdEnv indicates a production deployment environment.
	ProdEnv Environment = "production"
	// TestEnv indicates an automated test execution environment.
	TestEnv Environment = "test"
)

// Current inspects the ENV variable and returns the operational Environment, defaulting to DevEnv.
func Current() Environment {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("ENV"))) {
	case "production":
		return ProdEnv
	case "test":
		return TestEnv
	default:
		return DevEnv
	}
}

// ErrMissing is returned when a required environment variable is not defined or is empty.
var ErrMissing = errors.New("required environment variable not set")

// Require retrieves a required environment variable or returns ErrMissing if unset or empty.
func Require(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return "", fmt.Errorf("%s: %w", key, ErrMissing)
	}
	return val, nil
}

// String retrieves an environment variable or returns fallback if unset or empty.
func String(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	return val
}

// Int32 retrieves an environment variable parsed as an int32, or returns fallback if unset or empty.
func Int32(key string, fallback int32) (int32, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback, nil
	}

	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return fallback, fmt.Errorf("%s: %w", key, err)
	}

	return int32(n), nil
}

// Duration retrieves an environment variable parsed as a time.Duration, or returns fallback if unset or empty.
func Duration(key string, fallback time.Duration) (time.Duration, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback, nil
	}

	d, err := time.ParseDuration(val)
	if err != nil {
		return fallback, fmt.Errorf("%s: %w", key, err)
	}

	return d, nil
}

// Bool retrieves an environment variable parsed as a boolean, or returns fallback if unset or empty.
func Bool(key string, fallback bool) (bool, error) {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback, nil
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback, fmt.Errorf("%s: %w", key, err)
	}

	return b, nil
}
