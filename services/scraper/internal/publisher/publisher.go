// Package publisher provides NATS JetStream event publishing for scraper tasks.
package publisher

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/shared/nats"
)

// JetStreamPublisher publishes Protobuf messages to JetStream subjects.
type JetStreamPublisher interface {
	PublishProto(ctx context.Context, subject string, msg proto.Message) (*jetstream.PubAck, error)
}

// Publisher dispatches scraped page content and newly discovered URLs to NATS JetStream.
type Publisher struct {
	client JetStreamPublisher
}

// New creates and initializes a new Publisher instance.
func New(client JetStreamPublisher) *Publisher {
	return &Publisher{
		client: client,
	}
}

// PublishURLDiscovered emits a UrlDiscovered event notifying for URL discovery.
func (p *Publisher) PublishURLDiscovered(ctx context.Context, jobID, parentURL string, depth int32, newURLs []string) error {
	msg := &birdwatcherv1.UrlDiscovered{
		Meta:      nats.NewEventMetadata(jobID, nats.SubjectUrlDiscovered, parentURL),
		ParentUrl: parentURL,
		NewUrls:   newURLs,
		Depth:     depth,
	}

	if _, err := p.client.PublishProto(ctx, nats.SubjectUrlDiscovered, msg); err != nil {
		return fmt.Errorf("publish proto: %w", err)
	}

	return nil
}

// PublishPageScraped emits a PageScraped event containing the extracted content of a scraped page.
func (p *Publisher) PublishPageScraped(ctx context.Context, jobID, pageURL, title, bodyText string, imageURLs []string) error {
	msg := &birdwatcherv1.PageScraped{
		Meta:      nats.NewEventMetadata(jobID, nats.SubjectPageScraped, pageURL),
		Url:       pageURL,
		Title:     title,
		BodyText:  bodyText,
		ImageUrls: imageURLs,
	}

	if _, err := p.client.PublishProto(ctx, nats.SubjectPageScraped, msg); err != nil {
		return fmt.Errorf("publish proto: %w", err)
	}

	return nil
}
