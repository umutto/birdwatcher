package nats_test

import (
	"context"
	"errors"
	"testing"

	bwnats "github.com/umutto/birdwatcher/shared/nats"
)

func TestClient_CreateOrUpdateStream(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates stream with default config", func(t *testing.T) {
		t.Parallel()

		client := createTestClient(t)

		cfg := bwnats.DefaultStreamConfig()
		stream, err := client.CreateOrUpdateStream(context.Background(), cfg)

		if err != nil {
			t.Fatalf("expected stream to be created, got %v", err)
		}

		info, err := stream.Info(context.Background())
		if err != nil {
			t.Fatalf("expected stream info, got %v", err)
		}

		if info.Config.Name != cfg.Name {
			t.Errorf("stream name got %q, want %q", info.Config.Name, cfg.Name)
		}
	})

	t.Run("updates existing stream configuration idempotently", func(t *testing.T) {
		t.Parallel()

		client := createTestClient(t)

		cfg := bwnats.DefaultStreamConfig()
		if _, err := client.CreateOrUpdateStream(context.Background(), cfg); err != nil {
			t.Fatalf("expected initial stream creation to succeed, got %v", err)
		}

		cfg.Description = "Updated description"
		stream, err := client.CreateOrUpdateStream(context.Background(), cfg)
		if err != nil {
			t.Fatalf("expected stream update to succeed, got %v", err)
		}

		info, err := stream.Info(context.Background())
		if err != nil {
			t.Fatalf("expected stream info, got %v", err)
		}

		if info.Config.Description != cfg.Description {
			t.Errorf("stream description got %q, want %q", info.Config.Description, cfg.Description)
		}
	})

	t.Run("fails when context is canceled", func(t *testing.T) {
		t.Parallel()

		client := createTestClient(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := client.CreateOrUpdateStream(ctx, bwnats.DefaultStreamConfig()); !errors.Is(err, context.Canceled) {
			t.Errorf("expected error wrapping context.Canceled, got %v", err)
		}
	})

	t.Run("fails when client or jetstream is uninitialized", func(t *testing.T) {
		t.Parallel()

		client := &bwnats.Client{}

		if _, err := client.CreateOrUpdateStream(context.Background(), bwnats.DefaultStreamConfig()); err == nil {
			t.Errorf("expected error when client is uninitialized, got nil")
		}
	})
}

func createTestClient(t *testing.T) *bwnats.Client {
	t.Helper()

	url := startTestServer(t)
	clientCfg := bwnats.DefaultConfig()
	clientCfg.URL = url

	client, err := bwnats.Connect(context.Background(), clientCfg, nil)
	if err != nil {
		t.Fatalf("expected client to connect, got %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client

}
