package publisher_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/frontier/internal/publisher"
	"github.com/umutto/birdwatcher/shared/nats"
)

// mockJetStreamPublisher records the last published subject and message for testing.
type mockJetStreamPublisher struct {
	err     error
	subject string
	msg     proto.Message
}

func (m *mockJetStreamPublisher) PublishProto(_ context.Context, subject string, msg proto.Message) (*jetstream.PubAck, error) {
	m.subject = subject
	m.msg = msg
	return &jetstream.PubAck{}, m.err
}

func TestPublisher_PublishFetch(t *testing.T) {
	t.Parallel()

	const (
		testURL   = "https://audubon.org/birds"
		testDepth = int32(2)
	)
	testJobID := uuid.New().String()

	t.Run("successfully publishes page fetch request", func(t *testing.T) {
		t.Parallel()

		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		err := pub.PublishFetch(context.Background(), testJobID, testURL, testDepth)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.subject != nats.SubjectPageFetchRequested {
			t.Errorf("subject = %q, want %q", client.subject, nats.SubjectPageFetchRequested)
		}

		req, ok := client.msg.(*birdwatcherv1.PageFetchRequested)
		if !ok {
			t.Fatalf("expected *birdwatcherv1.PageFetchRequested, got %T", client.msg)
		}

		if req.GetUrl() != testURL {
			t.Errorf("url = %q, want %q", req.GetUrl(), testURL)
		}
		if req.GetDepth() != testDepth {
			t.Errorf("depth = %d, want %d", req.GetDepth(), testDepth)
		}
		if req.GetMeta().GetJobId() != testJobID {
			t.Errorf("job_id = %q, want %q", req.GetMeta().GetJobId(), testJobID)
		}
		if req.GetMeta().GetEventId() == "" {
			t.Error("expected non-empty event_id")
		}
		if req.GetMeta().GetTimestamp() == nil {
			t.Error("expected non-nil timestamp")
		}
	})

	t.Run("propagates publish error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("connection lost")
		client := &mockJetStreamPublisher{err: wantErr}
		pub := publisher.New(client)

		err := pub.PublishFetch(context.Background(), testJobID, testURL, testDepth)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error wrapping %v, got %v", wantErr, err)
		}
	})
}

func TestPublisher_PublishJobCompleted(t *testing.T) {
	t.Parallel()

	testJobID := uuid.New().String()

	t.Run("publishes completed status when error message is empty", func(t *testing.T) {
		t.Parallel()

		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		err := pub.PublishJobCompleted(context.Background(), testJobID, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.subject != nats.SubjectJobCompleted {
			t.Errorf("subject = %q, want %q", client.subject, nats.SubjectJobCompleted)
		}

		evt, ok := client.msg.(*birdwatcherv1.JobCompleted)
		if !ok {
			t.Fatalf("expected *birdwatcherv1.JobCompleted, got %T", client.msg)
		}

		if evt.GetStatus() != birdwatcherv1.CrawlStatus_CRAWL_STATUS_COMPLETED {
			t.Errorf("status = %v, want %v", evt.GetStatus(), birdwatcherv1.CrawlStatus_CRAWL_STATUS_COMPLETED)
		}
		if evt.GetErrorMessage() != "" {
			t.Errorf("expected empty error_message, got %q", evt.GetErrorMessage())
		}
		if evt.GetMeta().GetJobId() != testJobID {
			t.Errorf("job_id = %q, want %q", evt.GetMeta().GetJobId(), testJobID)
		}
	})

	t.Run("publishes failed status with error message when populated", func(t *testing.T) {
		t.Parallel()

		const failReason = "crawl depth exceeded limit"
		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		err := pub.PublishJobCompleted(context.Background(), testJobID, failReason)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		evt, ok := client.msg.(*birdwatcherv1.JobCompleted)
		if !ok {
			t.Fatalf("expected *birdwatcherv1.JobCompleted, got %T", client.msg)
		}

		if evt.GetStatus() != birdwatcherv1.CrawlStatus_CRAWL_STATUS_FAILED {
			t.Errorf("status = %v, want %v", evt.GetStatus(), birdwatcherv1.CrawlStatus_CRAWL_STATUS_FAILED)
		}
		if evt.GetErrorMessage() != failReason {
			t.Errorf("error_message = %q, want %q", evt.GetErrorMessage(), failReason)
		}
	})

	t.Run("propagates publish error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("timeout publishing job completed")
		client := &mockJetStreamPublisher{err: wantErr}
		pub := publisher.New(client)

		err := pub.PublishJobCompleted(context.Background(), testJobID, "")
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error wrapping %v, got %v", wantErr, err)
		}
	})
}
