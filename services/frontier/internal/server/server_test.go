package server_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/frontier/internal/config"
	"github.com/umutto/birdwatcher/services/frontier/internal/engine"
	"github.com/umutto/birdwatcher/services/frontier/internal/server"
)

// mockStarter implements server.Starter for testing.
type mockStarter struct {
	err error
}

// StartJob returns the configured mock error.
func (m *mockStarter) StartJob(_ context.Context, _, _ string, _, _ int32) error {
	return m.err
}

func TestServer_StartJob(t *testing.T) {
	t.Parallel()

	const validJobID = "550e8400-e29b-41d4-a716-446655440000"
	const validURL = "https://audubon.org"

	tests := []struct {
		name     string
		req      *birdwatcherv1.StartJobRequest
		starter  server.Starter
		wantCode codes.Code
	}{
		{
			name: "valid request",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  validURL,
				MaxDepth: 2,
				MaxPages: 50,
			},
			wantCode: codes.OK,
		},
		{
			name: "invalid job_id not a uuid",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    "not-a-uuid",
				RootUrl:  validURL,
				MaxDepth: 2,
				MaxPages: 50,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "invalid root_url empty",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  "",
				MaxDepth: 2,
				MaxPages: 50,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "invalid root_url relative path",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  "/path/to/birds",
				MaxDepth: 2,
				MaxPages: 50,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "invalid root_url unsupported scheme",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  "ftp://example.com",
				MaxDepth: 2,
				MaxPages: 50,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "invalid max_depth zero",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  validURL,
				MaxDepth: 0,
				MaxPages: 50,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "invalid max_pages zero",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  validURL,
				MaxDepth: 2,
				MaxPages: 0,
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "duplicate job error",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  validURL,
				MaxDepth: 2,
				MaxPages: 50,
			},
			starter:  &mockStarter{err: engine.ErrJobAlreadyExists},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "internal engine failure",
			req: &birdwatcherv1.StartJobRequest{
				JobId:    validJobID,
				RootUrl:  validURL,
				MaxDepth: 2,
				MaxPages: 50,
			},
			starter:  &mockStarter{err: errors.New("valkey store failed")},
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			starter := tt.starter
			if starter == nil {
				starter = &mockStarter{}
			}

			s := server.New(
				slog.New(slog.DiscardHandler),
				config.ServerConfig{Port: 9090},
				starter,
			)

			resp, err := s.StartJob(context.Background(), tt.req)
			if status.Code(err) != tt.wantCode {
				t.Fatalf("expected code %v, got %v (err: %v)", tt.wantCode, status.Code(err), err)
			}

			if tt.wantCode == codes.OK && resp == nil {
				t.Fatal("expected non-nil response for successful request")
			}
		})
	}
}
