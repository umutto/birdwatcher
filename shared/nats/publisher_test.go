package nats_test

import (
	"context"
	"errors"
	"testing"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	bwnats "github.com/umutto/birdwatcher/shared/nats"
)

func TestClient_PublishProto(t *testing.T) {
	t.Parallel()

	testSubject := bwnats.SubjectPageFetchRequested
	testEvent := &birdwatcherv1.PageFetchRequested{
		Meta: &birdwatcherv1.EventMetadata{
			JobId: "job-123",
		},
		Url:   "https://example.com",
		Depth: 1,
	}

	t.Run("successfully publishes proto message to jetstream", func(t *testing.T) {
		t.Parallel()

		client := createTestClientWithStream(t)

		ack, err := client.PublishProto(context.Background(), testSubject, testEvent)
		if err != nil {
			t.Fatalf("publish proto to %q: %v", testSubject, err)
		}

		if ack.Stream != bwnats.DefaultStreamName {
			t.Errorf("ack.Stream got %q, want %q", ack.Stream, bwnats.DefaultStreamName)
		}

		if ack.Sequence == 0 {
			t.Errorf("ack.Sequence got 0, want > 0")
		}
	})

	t.Run("fails when publishing nil message", func(t *testing.T) {
		t.Parallel()

		client := createTestClientWithStream(t)

		_, err := client.PublishProto(context.Background(), testSubject, nil)
		if err == nil {
			t.Errorf("expected error publishing nil message, got nil")
		}
	})

	t.Run("fails when context is canceled", func(t *testing.T) {
		t.Parallel()

		client := &bwnats.Client{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := client.PublishProto(ctx, testSubject, testEvent)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected error wrapping context.Canceled, got %v", err)
		}
	})

	t.Run("fails when client or jetstream is uninitialized", func(t *testing.T) {
		t.Parallel()

		var nilClient *bwnats.Client
		if _, err := nilClient.PublishProto(context.Background(), testSubject, testEvent); err == nil {
			t.Errorf("expected error on nil client pointer, got nil")
		}

		emptyClient := &bwnats.Client{}
		if _, err := emptyClient.PublishProto(context.Background(), testSubject, testEvent); err == nil {
			t.Errorf("expected error on uninitialized client, got nil")
		}
	})
}

func createTestClientWithStream(t *testing.T) *bwnats.Client {
	t.Helper()

	client := createTestClient(t)
	if _, err := client.CreateOrUpdateStream(context.Background(), bwnats.DefaultStreamConfig()); err != nil {
		t.Fatalf("failed to setup stream: %v", err)
	}

	return client
}
