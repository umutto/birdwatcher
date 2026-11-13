// Package consumer provides NATS JetStream event handlers for the scraper service.
package consumer

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/scraper/internal/parser"
	"github.com/umutto/birdwatcher/shared/nats"
)

const (
	defaultConsumerName = "scraper-consumer"
)

var filterSubjects = []string{
	nats.SubjectPageFetchRequested,
}

// PageFetcher retrieves raw HTML bytes from a target URL.
type PageFetcher interface {
	Fetch(ctx context.Context, targetURL string) ([]byte, error)
}

// EventPublisher dispatches scraped page content and discovered links to JetStream.
type EventPublisher interface {
	PublishURLDiscovered(ctx context.Context, jobID, parentURL string, depth int32, newURLs []string) error
	PublishPageScraped(ctx context.Context, jobID, pageURL, title, bodyText string, imageURLs []string) error
}

// Consumer processes PageFetchRequested events from NATS JetStream.
type Consumer struct {
	log       *slog.Logger
	fetcher   PageFetcher
	publisher EventPublisher
	js        jetstream.JetStream
}

// New creates a new JetStream event consumer.
func New(log *slog.Logger, fetcher PageFetcher, publisher EventPublisher, js jetstream.JetStream) *Consumer {
	return &Consumer{
		log:       log,
		fetcher:   fetcher,
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
	case nats.SubjectPageFetchRequested:
		if err := c.handlePageFetchRequested(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle page fetch requested", slog.Any("error", err))
		}
	default:
		c.log.WarnContext(ctx, "unknown subject", slog.String("subject", msg.Subject()))
		_ = msg.Term()
	}
}

// handlePageFetchRequested fetches the target URL, parses its HTML, and publishes the results.
func (c *Consumer) handlePageFetchRequested(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.PageFetchRequested
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	jobID, err := uuid.Parse(evt.GetMeta().GetJobId())
	if err != nil {
		_ = msg.Term()
		return fmt.Errorf("parse uuid: %w", err)
	}

	body, err := c.fetcher.Fetch(ctx, evt.GetUrl())
	if err != nil {
		_ = msg.Nak()
		return fmt.Errorf("fetch: %w", err)
	}

	res, err := parser.Parse(evt.GetUrl(), bytes.NewReader(body))
	if err != nil {
		_ = msg.Term()
		return fmt.Errorf("parse: %w", err)
	}

	if err := c.publisher.PublishPageScraped(ctx, jobID.String(), evt.GetUrl(), res.Title, res.BodyText, res.ImageURLs); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("publish page scraped: %w", err)
	}

	if err := c.publisher.PublishURLDiscovered(ctx, jobID.String(), evt.GetUrl(), evt.GetDepth(), res.Links); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("publish url discovered: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}
