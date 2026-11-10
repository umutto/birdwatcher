package consumer_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/frontier/internal/consumer"
	"github.com/umutto/birdwatcher/shared/nats"
)

// mockURLHandler records call arguments and returns a configured error.
type mockURLHandler struct {
	err error

	jobID          string
	parentURL      string
	parentDepth    int32
	discoveredURLs []string
}

// HandleDiscoveredURLs captures arguments and returns the configured mock error.
func (m *mockURLHandler) HandleDiscoveredURLs(
	_ context.Context,
	jobID, parentURL string,
	parentDepth int32,
	discoveredURLs []string,
) error {
	m.jobID = jobID
	m.parentURL = parentURL
	m.parentDepth = parentDepth
	m.discoveredURLs = discoveredURLs
	return m.err
}

// mockMsg implements jetstream.Msg for unit testing.
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
	validEvt := &birdwatcherv1.UrlDiscovered{
		Meta: &birdwatcherv1.EventMetadata{
			JobId: validJobID,
		},
		ParentUrl: "https://audubon.org/birds",
		NewUrls:   []string{"https://audubon.org/birds/puffin", "https://audubon.org/birds/robin"},
		Depth:     1,
	}
	validData, err := proto.Marshal(validEvt)
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}

	tests := []struct {
		name       string
		subject    string
		data       []byte
		handlerErr error
		wantAck    bool
		wantNak    bool
		wantTerm   bool
	}{
		{
			name:    "valid message acks and forwards to handler",
			subject: nats.SubjectUrlDiscovered,
			data:    validData,
			wantAck: true,
		},
		{
			name:     "malformed protobuf terms message",
			subject:  nats.SubjectUrlDiscovered,
			data:     []byte("not-a-valid-proto"),
			wantTerm: true,
		},
		{
			name:    "invalid job_id terms message",
			subject: nats.SubjectUrlDiscovered,
			data: func() []byte {
				b, _ := proto.Marshal(&birdwatcherv1.UrlDiscovered{
					Meta: &birdwatcherv1.EventMetadata{JobId: "not-a-uuid"},
				})
				return b
			}(),
			wantTerm: true,
		},
		{
			name:       "handler error naks message",
			subject:    nats.SubjectUrlDiscovered,
			data:       validData,
			handlerErr: errors.New("valkey timeout"),
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

			handler := &mockURLHandler{err: tt.handlerErr}
			c := consumer.New(
				slog.New(slog.DiscardHandler),
				handler,
				nil,
			)

			msg := &mockMsg{
				subject: tt.subject,
				data:    tt.data,
			}

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
				if handler.jobID != validJobID {
					t.Errorf("jobID = %q, want %q", handler.jobID, validJobID)
				}
				if handler.parentURL != "https://audubon.org/birds" {
					t.Errorf("parentURL = %q, want https://audubon.org/birds", handler.parentURL)
				}
				if handler.parentDepth != 1 {
					t.Errorf("parentDepth = %d, want 1", handler.parentDepth)
				}
				if len(handler.discoveredURLs) != 2 {
					t.Errorf("discoveredURLs length = %d, want 2", len(handler.discoveredURLs))
				}
			}
		})
	}
}
