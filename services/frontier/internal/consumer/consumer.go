// Package consumer provides NATS JetStream event handlers for frontier URL discovery.
package consumer

import (
	"context"
	"fmt"
	"log/slog"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/shared/nats"
)

const (
	defaultConsumerName = "frontier-consumer"
)

var filterSubjects = []string{
	nats.SubjectUrlDiscovered,
}

// DiscoveredURLHandler processes newly discovered links for an active crawl job.
type DiscoveredURLHandler interface {
	HandleDiscoveredURLs(
		ctx context.Context,
		jobID,
		parentURL string,
		parentDepth int32,
		discoveredURLs []string,
	) error
}

// Consumer processes UrlDiscovered events from NATS JetStream and schedules them in the frontier.
type Consumer struct {
	log     *slog.Logger
	handler DiscoveredURLHandler
	js      jetstream.JetStream
}

// New creates a new JetStream event consumer.
func New(log *slog.Logger, handler DiscoveredURLHandler, js jetstream.JetStream) *Consumer {
	return &Consumer{
		log:     log,
		handler: handler,
		js:      js,
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
	case nats.SubjectUrlDiscovered:
		if err := c.handleURLDiscovered(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle url discovered", slog.Any("error", err))
		}
	default:
		c.log.WarnContext(ctx, "unknown subject", slog.String("subject", msg.Subject()))
		_ = msg.Term()
	}
}

func (c *Consumer) handleURLDiscovered(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.UrlDiscovered
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	jobID, err := uuid.Parse(evt.GetMeta().GetJobId())
	if err != nil {
		_ = msg.Term()
		return fmt.Errorf("parse uuid: %w", err)
	}

	if err = c.handler.HandleDiscoveredURLs(ctx, jobID.String(), evt.GetParentUrl(), evt.GetDepth(), evt.GetNewUrls()); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("handle discovered urls: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}
