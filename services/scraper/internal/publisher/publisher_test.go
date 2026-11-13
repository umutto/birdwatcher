package publisher_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/scraper/internal/publisher"
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

func TestPublisher_PublishURLDiscovered(t *testing.T) {
	t.Parallel()

	const (
		testParentURL = "https://example.com/birds"
		testDepth     = int32(1)
	)
	testNewURLs := []string{"https://example.com/birds/robin", "https://example.com/birds/eagle"}
	testJobID := uuid.New().String()

	t.Run("successfully publishes url discovered event", func(t *testing.T) {
		t.Parallel()

		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		err := pub.PublishURLDiscovered(context.Background(), testJobID, testParentURL, testDepth, testNewURLs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.subject != nats.SubjectUrlDiscovered {
			t.Errorf("subject = %q, want %q", client.subject, nats.SubjectUrlDiscovered)
		}

		evt, ok := client.msg.(*birdwatcherv1.UrlDiscovered)
		if !ok {
			t.Fatalf("expected *birdwatcherv1.UrlDiscovered, got %T", client.msg)
		}

		if evt.GetParentUrl() != testParentURL {
			t.Errorf("parent_url = %q, want %q", evt.GetParentUrl(), testParentURL)
		}
		if evt.GetDepth() != testDepth {
			t.Errorf("depth = %d, want %d", evt.GetDepth(), testDepth)
		}
		if !slices.Equal(evt.GetNewUrls(), testNewURLs) {
			t.Errorf("new_urls = %v, want %v", evt.GetNewUrls(), testNewURLs)
		}
		if evt.GetMeta().GetJobId() != testJobID {
			t.Errorf("job_id = %q, want %q", evt.GetMeta().GetJobId(), testJobID)
		}
		if evt.GetMeta().GetEventId() == "" {
			t.Error("expected non-empty event_id")
		}
		if evt.GetMeta().GetTimestamp() == nil {
			t.Error("expected non-nil timestamp")
		}
	})

	t.Run("propagates publish error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("connection failed")
		client := &mockJetStreamPublisher{err: wantErr}
		pub := publisher.New(client)

		err := pub.PublishURLDiscovered(context.Background(), testJobID, testParentURL, testDepth, testNewURLs)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error wrapping %v, got %v", wantErr, err)
		}
	})
}

func TestPublisher_PublishPageScraped(t *testing.T) {
	t.Parallel()

	const (
		testURL      = "https://example.com/birds/robin"
		testTitle    = "American Robin"
		testBodyText = "The American robin is a migratory songbird."
	)
	testImageURLs := []string{"https://example.com/images/robin1.jpg", "https://example.com/images/robin2.jpg"}
	testJobID := uuid.New().String()

	t.Run("successfully publishes page scraped event", func(t *testing.T) {
		t.Parallel()

		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		err := pub.PublishPageScraped(context.Background(), testJobID, testURL, testTitle, testBodyText, testImageURLs)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.subject != nats.SubjectPageScraped {
			t.Errorf("subject = %q, want %q", client.subject, nats.SubjectPageScraped)
		}

		evt, ok := client.msg.(*birdwatcherv1.PageScraped)
		if !ok {
			t.Fatalf("expected *birdwatcherv1.PageScraped, got %T", client.msg)
		}

		if evt.GetUrl() != testURL {
			t.Errorf("url = %q, want %q", evt.GetUrl(), testURL)
		}
		if evt.GetTitle() != testTitle {
			t.Errorf("title = %q, want %q", evt.GetTitle(), testTitle)
		}
		if evt.GetBodyText() != testBodyText {
			t.Errorf("body_text = %q, want %q", evt.GetBodyText(), testBodyText)
		}
		if !slices.Equal(evt.GetImageUrls(), testImageURLs) {
			t.Errorf("image_urls = %v, want %v", evt.GetImageUrls(), testImageURLs)
		}
		if evt.GetMeta().GetJobId() != testJobID {
			t.Errorf("job_id = %q, want %q", evt.GetMeta().GetJobId(), testJobID)
		}
		if evt.GetMeta().GetEventId() == "" {
			t.Error("expected non-empty event_id")
		}
		if evt.GetMeta().GetTimestamp() == nil {
			t.Error("expected non-nil timestamp")
		}
	})

	t.Run("propagates publish error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("timeout publishing page scraped")
		client := &mockJetStreamPublisher{err: wantErr}
		pub := publisher.New(client)

		err := pub.PublishPageScraped(context.Background(), testJobID, testURL, testTitle, testBodyText, testImageURLs)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error wrapping %v, got %v", wantErr, err)
		}
	})
}
