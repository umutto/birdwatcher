// Package server provides gRPC transport and lifecycle management for the frontier service.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"
	"uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/frontier/internal/config"
)

const (
	defaultShutdownTimeout = 15 * time.Second
)

// Server implements the gRPC FrontierServiceServer interface.
type Server struct {
	birdwatcherv1.UnimplementedFrontierServiceServer

	log *slog.Logger
	cfg config.ServerConfig
}

// New creates and initializes a new frontier gRPC Server instance.
func New(log *slog.Logger, cfg config.ServerConfig) *Server {
	return &Server{
		log: log,
		cfg: cfg,
	}
}

// Start binds the TCP listener, registers the gRPC service, and manages graceful shutdown.
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", s.cfg.Port, err)
	}
	defer func() {
		_ = lis.Close()
	}()

	grpcServer := grpc.NewServer()
	birdwatcherv1.RegisterFrontierServiceServer(grpcServer, s)

	errCh := make(chan error, 1)

	go func() {
		s.log.InfoContext(ctx, "starting grpc server", slog.Int("port", int(s.cfg.Port)))
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("grpc server failed: %w", err)
	case <-ctx.Done():
		s.log.InfoContext(ctx, "shutting down grpc server")
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		s.log.InfoContext(ctx, "grpc server stopped gracefully")
	case <-time.After(defaultShutdownTimeout):
		s.log.WarnContext(ctx, "grpc server shutdown timed out; forcing stop")
		grpcServer.Stop()
	}

	return nil
}

// StartJob validates and dispatches a new crawl job to the frontier queue.
func (s *Server) StartJob(
	ctx context.Context,
	req *birdwatcherv1.StartJobRequest,
) (*birdwatcherv1.StartJobResponse, error) {
	jobID, err := uuid.Parse(req.GetJobId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid job_id: must be a valid UUID")
	}

	rootURL, err := url.Parse(strings.TrimSpace(req.GetRootUrl()))
	if err != nil || (rootURL.Scheme != "http" && rootURL.Scheme != "https") || rootURL.Host == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid root_url: must be an absolute http/https URL")
	}

	maxDepth := req.GetMaxDepth()
	if maxDepth <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid max_depth: must be greater than 0")
	}

	maxPages := req.GetMaxPages()
	if maxPages <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid max_pages: must be greater than 0")
	}

	s.log.InfoContext(
		ctx,
		"start job",
		slog.String("job_id", jobID.String()),
		slog.String("root_url", rootURL.String()),
		slog.Int("max_depth", int(maxDepth)),
		slog.Int("max_pages", int(maxPages)),
	)

	return &birdwatcherv1.StartJobResponse{}, nil
}
