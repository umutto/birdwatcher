package nats_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	bwnats "github.com/umutto/birdwatcher/shared/nats"
)

func TestConnect(t *testing.T) {
	t.Run("successfully connects and closes with live JetStream broker", func(t *testing.T) {
		url := startTestServer(t)

		cfg := bwnats.DefaultConfig()
		cfg.URL = url

		client, err := bwnats.Connect(context.Background(), cfg, nil)
		if err != nil {
			t.Fatalf("expected successful connection, got %v", err)
		}

		if !client.Conn.IsConnected() {
			t.Errorf("expected connection to be open")
		}
		if client.JS == nil {
			t.Errorf("expected non-nil JetStream context")
		}

		if err := client.Close(); err != nil {
			t.Errorf("expected clean close, got %v", err)
		}
		if !client.Conn.IsClosed() && !client.Conn.IsDraining() {
			t.Errorf("expected connection to be closed or draining")
		}
	})

	t.Run("fails on unreachable address", func(t *testing.T) {
		cfg := bwnats.Config{
			URL:     "nats://127.0.0.1:59999",
			Timeout: 5 * time.Millisecond,
		}

		client, err := bwnats.Connect(context.Background(), cfg, nil)

		if client != nil {
			t.Errorf("expected nil client on failed dial, got %+v", client)
		}
		if err == nil {
			t.Errorf("expected error connecting to unreachable address, got nil")
		}
	})

	t.Run("fails when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		client, err := bwnats.Connect(ctx, bwnats.DefaultConfig(), nil)

		if client != nil {
			t.Errorf("expected nil client, got %+v", client)
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected error wrapping context.Canceled, got %v", err)
		}
	})
}

func TestClient_Close(t *testing.T) {
	t.Run("safe with nil connection", func(t *testing.T) {
		client := &bwnats.Client{}
		if err := client.Close(); err != nil {
			t.Errorf("expected nil error on nil Conn close, got %v", err)
		}

	})
}

func startTestServer(t *testing.T) string {
	t.Helper()

	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
	}

	s, err := server.NewServer(opts)
	if err != nil {
		t.Fatalf("failed to create test NATS server: %v", err)
	}

	go s.Start()

	if !s.ReadyForConnections(5 * time.Second) {
		t.Fatal("test NATS server timed out waiting to be ready")
	}

	// Graceful shutdown after test completes
	t.Cleanup(func() {
		s.Shutdown()
	})

	return s.ClientURL()
}
