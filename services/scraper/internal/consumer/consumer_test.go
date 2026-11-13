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
	"github.com/umutto/birdwatcher/services/scraper/internal/consumer"
	"github.com/umutto/birdwatcher/shared/nats"
)

type mockFetcher struct {
	body []byte
	err  error
}

func (m *mockFetcher) Fetch(_ context.Context, _ string) ([]byte, error) {
	return m.body, m.err
}

type mockPublisher struct {
	scrapedErr    error
	discoveredErr error

	jobID      string
	pageURL    string
	title      string
	bodyText   string
	imageURLs  []string
	parentURL  string
	depth      int32
	discovered []string
}

func (m *mockPublisher) PublishPageScraped(
	_ context.Context,
	jobID, pageURL, title, bodyText string,
	imageURLs []string,
) error {
	m.jobID = jobID
	m.pageURL = pageURL
	m.title = title
	m.bodyText = bodyText
	m.imageURLs = imageURLs
	return m.scrapedErr
}

func (m *mockPublisher) PublishURLDiscovered(
	_ context.Context,
	jobID, parentURL string,
	depth int32,
	newURLs []string,
) error {
	m.jobID = jobID
	m.parentURL = parentURL
	m.depth = depth
	m.discovered = newURLs
	return m.discoveredErr
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
	const validURL = "https://audubon.org/birds"
	validHTML := []byte(`<!DOCTYPE html><html><head><title>Birds</title></head><body><p>Hello</p><a href="/robin">Robin</a><img src="/puffin.jpg"></body></html>`)

	validEvt := &birdwatcherv1.PageFetchRequested{
		Meta:  &birdwatcherv1.EventMetadata{JobId: validJobID},
		Url:   validURL,
		Depth: 1,
	}
	validData, err := proto.Marshal(validEvt)
	if err != nil {
		t.Fatalf("marshal proto: %v", err)
	}

	tests := []struct {
		name          string
		subject       string
		data          []byte
		fetchErr      error
		scrapedErr    error
		discoveredErr error
		wantAck       bool
		wantNak       bool
		wantTerm      bool
	}{
		{
			name:    "valid message fetches, parses, publishes, and acks",
			subject: nats.SubjectPageFetchRequested,
			data:    validData,
			wantAck: true,
		},
		{
			name:     "malformed protobuf terms message",
			subject:  nats.SubjectPageFetchRequested,
			data:     []byte("not-a-valid-proto"),
			wantTerm: true,
		},
		{
			name:    "invalid job_id terms message",
			subject: nats.SubjectPageFetchRequested,
			data: func() []byte {
				b, _ := proto.Marshal(&birdwatcherv1.PageFetchRequested{
					Meta: &birdwatcherv1.EventMetadata{JobId: "not-a-uuid"},
					Url:  validURL,
				})
				return b
			}(),
			wantTerm: true,
		},
		{
			name:    "invalid base url terms message",
			subject: nats.SubjectPageFetchRequested,
			data: func() []byte {
				b, _ := proto.Marshal(&birdwatcherv1.PageFetchRequested{
					Meta: &birdwatcherv1.EventMetadata{JobId: validJobID},
					Url:  "://invalid-url",
				})
				return b
			}(),
			wantTerm: true,
		},
		{
			name:     "fetcher error naks message",
			subject:  nats.SubjectPageFetchRequested,
			data:     validData,
			fetchErr: errors.New("connection timeout"),
			wantNak:  true,
		},
		{
			name:       "publish page scraped error naks message",
			subject:    nats.SubjectPageFetchRequested,
			data:       validData,
			scrapedErr: errors.New("nats timeout"),
			wantNak:    true,
		},
		{
			name:          "publish url discovered error naks message",
			subject:       nats.SubjectPageFetchRequested,
			data:          validData,
			discoveredErr: errors.New("nats timeout"),
			wantNak:       true,
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

			fetcher := &mockFetcher{body: validHTML, err: tt.fetchErr}
			pub := &mockPublisher{scrapedErr: tt.scrapedErr, discoveredErr: tt.discoveredErr}
			c := consumer.New(slog.New(slog.DiscardHandler), fetcher, pub, nil)

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
				if pub.jobID != validJobID {
					t.Errorf("jobID = %q, want %q", pub.jobID, validJobID)
				}
				if pub.title != "Birds" {
					t.Errorf("title = %q, want Birds", pub.title)
				}
				if pub.bodyText != "Hello Robin" {
					t.Errorf("bodyText = %q, want %q", pub.bodyText, "Hello Robin")
				}
				if !slices.Equal(pub.discovered, []string{"https://audubon.org/robin"}) {
					t.Errorf("discovered = %v, want [https://audubon.org/robin]", pub.discovered)
				}
				if !slices.Equal(pub.imageURLs, []string{"https://audubon.org/puffin.jpg"}) {
					t.Errorf("imageURLs = %v, want [https://audubon.org/puffin.jpg]", pub.imageURLs)
				}
			}
		})
	}
}
