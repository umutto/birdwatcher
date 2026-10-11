// Package server provides HTTP transport, routing, and lifecycle management for the catalog service.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
)

// Pinger verifies connectivity to a backing dependency during readiness checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Server encapsulates HTTP server dependencies and route handling.
type Server struct {
	log     *slog.Logger
	repo    *repository.Queries
	db      Pinger
	cfg     config.ServerConfig
	handler http.Handler
}

// New creates and configures a new HTTP Server instance.
func New(log *slog.Logger, db Pinger, repo *repository.Queries, cfg config.ServerConfig) *Server {
	s := &Server{log: log, db: db, repo: repo, cfg: cfg}
	s.handler = s.routes()

	return s
}

// Start binds the HTTP server to the configured port and manages graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.Port),
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		s.log.InfoContext(ctx, "starting http server", slog.Int("port", int(s.cfg.Port)))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server failed: %w", err)
	case <-ctx.Done():
		s.log.InfoContext(ctx, "shutting down http server")
	}

	// ctx is already canceled; strip the cancel signal so srv.Shutdown doesn't abort immediately.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	return nil
}

// ServeHTTP dispatches HTTP requests to the server's internal router.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
