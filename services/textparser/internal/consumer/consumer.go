// Package consumer provides NATS JetStream event handlers for the textparser service.
package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/textparser/internal/matcher"
	"github.com/umutto/birdwatcher/shared/nats"
)

const (
	defaultConsumerName = "textparser-consumer"
)

var filterSubjects = []string{
	nats.SubjectPageScraped,
}

// TextMatcher scans page title and body text for Audubon bird species mentions.
type TextMatcher interface {
	Match(title, bodyText string) []matcher.Sighting
}

// EventPublisher dispatches detected text sightings to JetStream.
type EventPublisher interface {
	PublishSightingText(
		ctx context.Context,
		jobID, pageURL string,
		speciesPlateNumber int32,
		confidence float32,
		snippet string,
	) error
}

// Consumer processes PageScraped events from NATS JetStream.
type Consumer struct {
	log       *slog.Logger
	matcher   TextMatcher
	publisher EventPublisher
	js        jetstream.JetStream
}

// New creates a new JetStream event consumer.
func New(log *slog.Logger, textMatcher TextMatcher, publisher EventPublisher, js jetstream.JetStream) *Consumer {
	return &Consumer{
		log:       log,
		matcher:   textMatcher,
		publisher: publisher,
		js:        js,
	}
}

// Start provisions the durable JetStream consumer, begins processing pipeline events,
// and blocks until the context is canceled.
func (c *Consumer) Start(ctx context.Context) error {
	c.log.InfoContext(ctx, "starting jetstream consumer", slog.String("consumer", defaultConsumerName))
	cons, err := c.js.CreateOrUpdateConsumer(
		ctx,
		nats.DefaultStreamName,
		jetstream.ConsumerConfig{
			Durable:        defaultConsumerName,
			FilterSubjects: filterSubjects,
		},
	)
	if err != nil {
		return fmt.Errorf("create or update consumer: %w", err)
	}

	cc, err := cons.Consume(func(msg jetstream.Msg) { c.handleMessage(ctx, msg) })
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}
	defer cc.Stop()

	<-ctx.Done()
	c.log.InfoContext(ctx, "shutting down jetstream consumer", slog.String("consumer", defaultConsumerName))

	return nil
}

// handleMessage routes incoming JetStream messages to their respective event handlers.
func (c *Consumer) handleMessage(ctx context.Context, msg jetstream.Msg) {
	switch msg.Subject() {
	case nats.SubjectPageScraped:
		if err := c.handlePageScraped(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle page scraped", slog.Any("error", err))
		}
	default:
		c.log.WarnContext(ctx, "unknown subject", slog.String("subject", msg.Subject()))
		_ = msg.Term()
	}
}

// handlePageScraped scans scraped page text for species mentions and publishes sighting events.
func (c *Consumer) handlePageScraped(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.PageScraped
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	jobID, err := uuid.Parse(evt.GetMeta().GetJobId())
	if err != nil {
		_ = msg.Term()
		return fmt.Errorf("parse job_id: %w", err)
	}

	sightings := c.matcher.Match(evt.GetTitle(), evt.GetBodyText())
	for _, sighting := range sightings {
		if err := c.publisher.PublishSightingText(
			ctx,
			jobID.String(),
			evt.GetUrl(),
			sighting.SpeciesPlateNumber,
			sighting.Confidence,
			sighting.Snippet,
		); err != nil {
			_ = msg.Nak()
			return fmt.Errorf("publish: %w", err)
		}
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}
