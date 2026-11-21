package consumer_test

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/textparser/internal/consumer"
	"github.com/umutto/birdwatcher/services/textparser/internal/matcher"
	"github.com/umutto/birdwatcher/shared/nats"
)

type mockMatcher struct {
	sightings []matcher.Sighting
	title     string
	bodyText  string
}

func (m *mockMatcher) Match(title, bodyText string) []matcher.Sighting {
	m.title = title
	m.bodyText = bodyText
	return m.sightings
}

type mockPublisher struct {
	err       error
	jobID     string
	pageURL   string
	sightings []matcher.Sighting
}

func (m *mockPublisher) PublishSightingText(
	_ context.Context,
	jobID, pageURL string,
	speciesPlateNumber int32,
	confidence float32,
	snippet string,
) error {
	if m.err != nil {
		return m.err
	}
	m.jobID = jobID
	m.pageURL = pageURL
	m.sightings = append(m.sightings, matcher.Sighting{
		SpeciesPlateNumber: speciesPlateNumber,
		Confidence:         confidence,
		Snippet:            snippet,
	})
	return nil
}

type mockMsg struct {
	jetstream.Msg
	subject string
	data    []byte
	acked   bool
	naked   bool
	termed  bool
}

func (m *mockMsg) Subject() string { return m.subject }
func (m *mockMsg) Data() []byte    { return m.data }
func (m *mockMsg) Ack() error      { m.acked = true; return nil }
func (m *mockMsg) Nak() error      { m.naked = true; return nil }
func (m *mockMsg) Term() error     { m.termed = true; return nil }

func TestConsumer_HandleMessage(t *testing.T) {
	t.Parallel()

	validJobID := uuid.New().String()
	const (
		validURL   = "https://audubon.org/birds"
		validTitle = "Bald Eagle Guide"
		validBody  = "The Bald Eagle and Common Loon live near lakes."
	)

	validData, err := proto.Marshal(&birdwatcherv1.PageScraped{
		Meta:     &birdwatcherv1.EventMetadata{JobId: validJobID},
		Url:      validURL,
		Title:    validTitle,
		BodyText: validBody,
	})
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}

	sightings := []matcher.Sighting{
		{SpeciesPlateNumber: 11, Confidence: 1.0, Snippet: "bald eagle guide"},
		{SpeciesPlateNumber: 306, Confidence: 1.0, Snippet: "common loon live near lakes."},
	}

	tests := []struct {
		name       string
		subject    string
		data       []byte
		sightings  []matcher.Sighting
		publishErr error
		wantAck    bool
		wantNak    bool
		wantTerm   bool
	}{
		{
			name:      "valid message publishes sightings and acks",
			subject:   nats.SubjectPageScraped,
			data:      validData,
			sightings: sightings,
			wantAck:   true,
		},
		{
			name:    "no sightings acks without publishing",
			subject: nats.SubjectPageScraped,
			data:    validData,
			wantAck: true,
		},
		{
			name:     "malformed protobuf terms message",
			subject:  nats.SubjectPageScraped,
			data:     []byte("not-a-valid-proto"),
			wantTerm: true,
		},
		{
			name:    "invalid job_id terms message",
			subject: nats.SubjectPageScraped,
			data: func() []byte {
				b, _ := proto.Marshal(&birdwatcherv1.PageScraped{
					Meta: &birdwatcherv1.EventMetadata{JobId: "not-a-uuid"},
					Url:  validURL,
				})
				return b
			}(),
			wantTerm: true,
		},
		{
			name:       "publisher error naks message",
			subject:    nats.SubjectPageScraped,
			data:       validData,
			sightings:  sightings,
			publishErr: errors.New("nats timeout"),
			wantNak:    true,
		},
		{
			name:     "unknown subject terms message",
			subject:  "birdwatcher.v1.unknown.subject",
			data:     validData,
			wantTerm: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &mockMatcher{sightings: tt.sightings}
			pub := &mockPublisher{err: tt.publishErr}
			c := consumer.New(slog.New(slog.DiscardHandler), m, pub, nil)

			msg := &mockMsg{subject: tt.subject, data: tt.data}
			c.HandleMessage(context.Background(), msg)

			if msg.acked != tt.wantAck {
				t.Errorf("acked = %v, want %v", msg.acked, tt.wantAck)
			}
			if msg.naked != tt.wantNak {
				t.Errorf("naked = %v, want %v", msg.naked, tt.wantNak)
			}
			if msg.termed != tt.wantTerm {
				t.Errorf("termed = %v, want %v", msg.termed, tt.wantTerm)
			}

			if tt.wantAck {
				if m.title != validTitle || m.bodyText != validBody {
					t.Errorf("Match(%q, %q), want (%q, %q)", m.title, m.bodyText, validTitle, validBody)
				}
				if !slices.Equal(pub.sightings, tt.sightings) {
					t.Errorf("published sightings = %+v, want %+v", pub.sightings, tt.sightings)
				}
				if len(tt.sightings) > 0 && (pub.jobID != validJobID || pub.pageURL != validURL) {
					t.Errorf("published (jobID=%q, pageURL=%q), want (%q, %q)", pub.jobID, pub.pageURL, validJobID, validURL)
				}
			}
		})
	}
}
