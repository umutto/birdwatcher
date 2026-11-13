// Package publisher provides NATS JetStream event publishing for crawl traversal tasks.
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

// Publisher dispatches crawl task and lifecycle events to NATS JetStream.
type Publisher struct {
	client JetStreamPublisher
}

// New creates and initializes a new Publisher instance.
func New(client JetStreamPublisher) *Publisher {
	return &Publisher{
		client: client,
	}
}

// PublishFetch emits a PageFetchRequested event to instruct scraper workers to fetch a URL.
func (p *Publisher) PublishFetch(ctx context.Context, jobID, targetURL string, depth int32) error {
	msg := &birdwatcherv1.PageFetchRequested{
		Meta:  nats.NewEventMetadata(jobID, nats.SubjectPageFetchRequested, targetURL),
		Url:   targetURL,
		Depth: depth,
	}

	if _, err := p.client.PublishProto(ctx, nats.SubjectPageFetchRequested, msg); err != nil {
		return fmt.Errorf("publish proto: %w", err)
	}

	return nil
}

// PublishJobCompleted emits a JobCompleted event marking the termination of a crawl job.
func (p *Publisher) PublishJobCompleted(ctx context.Context, jobID, errMsg string) error {
	msg := &birdwatcherv1.JobCompleted{
		Meta:   nats.NewEventMetadata(jobID, nats.SubjectJobCompleted),
		Status: birdwatcherv1.CrawlStatus_CRAWL_STATUS_COMPLETED,
	}
	if errMsg != "" {
		msg.Status = birdwatcherv1.CrawlStatus_CRAWL_STATUS_FAILED
		msg.ErrorMessage = errMsg
	}

	if _, err := p.client.PublishProto(ctx, nats.SubjectJobCompleted, msg); err != nil {
		return fmt.Errorf("publish proto: %w", err)
	}

	return nil
}
