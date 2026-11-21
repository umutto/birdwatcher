// Package publisher provides NATS JetStream event publishing for textparser sightings.
package publisher

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/shared/nats"
)

// JetStreamPublisher publishes Protobuf messages to JetStream subjects.
type JetStreamPublisher interface {
	PublishProto(ctx context.Context, subject string, msg proto.Message) (*jetstream.PubAck, error)
}

// Publisher dispatches detected bird text sightings to NATS JetStream.
type Publisher struct {
	client JetStreamPublisher
}

// New creates and initializes a new Publisher instance.
func New(client JetStreamPublisher) *Publisher {
	return &Publisher{
		client: client,
	}
}

// PublishSightingText emits a SightingText event for a species matched on a scraped page.
func (p *Publisher) PublishSightingText(
	ctx context.Context,
	jobID, pageURL string,
	speciesPlateNumber int32,
	confidence float32,
	snippet string,
) error {
	plate := strconv.Itoa(int(speciesPlateNumber))
	msg := &birdwatcherv1.SightingText{
		Meta:               nats.NewEventMetadata(jobID, nats.SubjectSightingText, pageURL, plate),
		PageUrl:            pageURL,
		SpeciesPlateNumber: speciesPlateNumber,
		Confidence:         confidence,
		Snippet:            snippet,
	}

	if _, err := p.client.PublishProto(ctx, nats.SubjectSightingText, msg); err != nil {
		return fmt.Errorf("publish proto: %w", err)
	}

	return nil
}
