package publisher_test

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/textparser/internal/publisher"
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

func TestPublisher_PublishSightingText(t *testing.T) {
	t.Parallel()

	const (
		testPageURL     = "https://audubon.org/birds/bald-eagle"
		testPlate       = int32(11)
		testConfidence  = float32(1.0)
		testSnippet     = "spotted a bald eagle near the river"
		testMatchedText = "bald eagle"
	)
	testJobID := uuid.New().String()

	t.Run("successfully publishes sighting text event", func(t *testing.T) {
		t.Parallel()

		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		err := pub.PublishSightingText(context.Background(), testJobID, testPageURL, testPlate, testConfidence, testSnippet, testMatchedText)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if client.subject != nats.SubjectSightingText {
			t.Errorf("subject = %q, want %q", client.subject, nats.SubjectSightingText)
		}

		evt, ok := client.msg.(*birdwatcherv1.SightingText)
		if !ok {
			t.Fatalf("expected *birdwatcherv1.SightingText, got %T", client.msg)
		}

		if evt.GetPageUrl() != testPageURL {
			t.Errorf("page_url = %q, want %q", evt.GetPageUrl(), testPageURL)
		}
		if evt.GetSpeciesPlateNumber() != testPlate {
			t.Errorf("species_plate_number = %d, want %d", evt.GetSpeciesPlateNumber(), testPlate)
		}
		if evt.GetConfidence() != testConfidence {
			t.Errorf("confidence = %v, want %v", evt.GetConfidence(), testConfidence)
		}
		if evt.GetSnippet() != testSnippet {
			t.Errorf("snippet = %q, want %q", evt.GetSnippet(), testSnippet)
		}
		if evt.GetMatchedText() != testMatchedText {
			t.Errorf("matched_text = %q, want %q", evt.GetMatchedText(), testMatchedText)
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

	t.Run("generates deterministic per-species event ids", func(t *testing.T) {
		t.Parallel()

		client := &mockJetStreamPublisher{}
		pub := publisher.New(client)

		if err := pub.PublishSightingText(context.Background(), testJobID, testPageURL, 11, testConfidence, testSnippet, testMatchedText); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		firstID := client.msg.(*birdwatcherv1.SightingText).GetMeta().GetEventId()

		if err := pub.PublishSightingText(context.Background(), testJobID, testPageURL, 31, testConfidence, testSnippet, testMatchedText); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		otherPlateID := client.msg.(*birdwatcherv1.SightingText).GetMeta().GetEventId()

		if firstID == otherPlateID {
			t.Errorf("expected distinct event_ids for different species on the same page, both were %q", firstID)
		}

		if err := pub.PublishSightingText(context.Background(), testJobID, testPageURL, 11, testConfidence, testSnippet, testMatchedText); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		retryID := client.msg.(*birdwatcherv1.SightingText).GetMeta().GetEventId()

		if firstID != retryID {
			t.Errorf("expected deterministic event_id on retry: got %q, want %q", retryID, firstID)
		}
	})

	t.Run("propagates publish error", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("connection failed")
		client := &mockJetStreamPublisher{err: wantErr}
		pub := publisher.New(client)

		err := pub.PublishSightingText(context.Background(), testJobID, testPageURL, testPlate, testConfidence, testSnippet, testMatchedText)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error wrapping %v, got %v", wantErr, err)
		}
	})
}
