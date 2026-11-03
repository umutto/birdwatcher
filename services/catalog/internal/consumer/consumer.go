// Package consumer provides NATS JetStream event handlers for catalog updates.
package consumer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/shared/nats"
)

const (
	defaultConsumerName = "catalog-consumer"
)

var filterSubjects = []string{
	nats.SubjectPageScraped,
	nats.SubjectSightingText,
	nats.SubjectSightingImage,
	nats.SubjectJobCompleted,
}

// Consumer processes pipeline events from NATS JetStream and updates the database.
type Consumer struct {
	log  *slog.Logger
	repo repository.Querier
	js   jetstream.JetStream
}

// New creates a new JetStream event consumer.
func New(log *slog.Logger, repo repository.Querier, js jetstream.JetStream) *Consumer {
	return &Consumer{
		log:  log,
		repo: repo,
		js:   js,
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
	c.log.InfoContext(ctx, "shutting down jetstream consumer")

	return nil
}

// handleMessage routes incoming JetStream messages to their respective event handlers.
func (c *Consumer) handleMessage(ctx context.Context, msg jetstream.Msg) {
	switch msg.Subject() {
	case nats.SubjectPageScraped:
		if err := c.handlePageScraped(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle page scraped", slog.Any("error", err))
		}
	case nats.SubjectSightingText:
		if err := c.handleSightingText(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle sighting text", slog.Any("error", err))
		}
	case nats.SubjectSightingImage:
		if err := c.handleSightingImage(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle sighting image", slog.Any("error", err))
		}
	case nats.SubjectJobCompleted:
		if err := c.handleJobCompleted(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "handle job completed", slog.Any("error", err))
		}
	default:
		c.log.WarnContext(ctx, "unknown subject", slog.String("subject", msg.Subject()))
		_ = msg.Term()
	}
}

// handlePageScraped processes a PageScraped event by updating the crawl title and progress counters.
func (c *Consumer) handlePageScraped(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.PageScraped
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	crawlID, err := parseCrawlID(evt.Meta)
	if err != nil {
		_ = msg.Term()
		return err
	}

	if evt.Title != "" {
		if err := c.repo.UpdateCrawlTitle(ctx, repository.UpdateCrawlTitleParams{
			ID:    crawlID,
			Title: &evt.Title,
		}); err != nil {
			_ = msg.Nak()
			return fmt.Errorf("update crawl title: %w", err)
		}
	}

	if err := c.repo.IncrementCrawlProgress(ctx, repository.IncrementCrawlProgressParams{
		ID:                 crawlID,
		PagesCrawledCount:  1,
		ImagesScannedCount: int32(len(evt.ImageUrls)),
	}); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("increment crawl progress: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

// handleSightingText processes a SightingText event by creating a text-based sighting observation.
func (c *Consumer) handleSightingText(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.SightingText
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	crawlID, err := parseCrawlID(evt.Meta)
	if err != nil {
		_ = msg.Term()
		return err
	}

	if _, err = c.repo.CreateSighting(ctx, repository.CreateSightingParams{
		CrawlID:            crawlID,
		SpeciesPlateNumber: evt.SpeciesPlateNumber,
		SourceType:         repository.SightingSourceTypeText,
		Confidence:         evt.Confidence,
		PageUrl:            evt.PageUrl,
		Snippet:            &evt.Snippet,
		ImageUrl:           nil,
	}); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("create sighting: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

// handleSightingImage processes a SightingImage event by creating an image-based sighting observation.
func (c *Consumer) handleSightingImage(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.SightingImage
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	crawlID, err := parseCrawlID(evt.Meta)
	if err != nil {
		_ = msg.Term()
		return err
	}

	if _, err = c.repo.CreateSighting(ctx, repository.CreateSightingParams{
		CrawlID:            crawlID,
		SpeciesPlateNumber: evt.SpeciesPlateNumber,
		SourceType:         repository.SightingSourceTypeImage,
		Confidence:         evt.Confidence,
		PageUrl:            evt.PageUrl,
		Snippet:            nil,
		ImageUrl:           &evt.ImageUrl,
	}); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("create sighting: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

// handleJobCompleted processes a JobCompleted event by transitioning the crawl to its terminal state.
func (c *Consumer) handleJobCompleted(ctx context.Context, msg jetstream.Msg) error {
	var evt birdwatcherv1.JobCompleted
	if err := proto.Unmarshal(msg.Data(), &evt); err != nil {
		_ = msg.Term()
		return fmt.Errorf("unmarshal: %w", err)
	}

	crawlID, err := parseCrawlID(evt.Meta)
	if err != nil {
		_ = msg.Term()
		return err
	}

	var state repository.CrawlState
	switch evt.Status {
	case birdwatcherv1.CrawlStatus_CRAWL_STATUS_COMPLETED:
		state = repository.CrawlStateCompleted
	case birdwatcherv1.CrawlStatus_CRAWL_STATUS_FAILED:
		state = repository.CrawlStateFailed
	case birdwatcherv1.CrawlStatus_CRAWL_STATUS_CANCELLED:
		state = repository.CrawlStateCancelled
	default:
		_ = msg.Term()
		return fmt.Errorf("unsupported crawl status: %v", evt.Status)
	}

	if _, err = c.repo.UpdateCrawlState(ctx, repository.UpdateCrawlStateParams{
		State: state,
		ID:    crawlID,
	}); err != nil {
		_ = msg.Nak()
		return fmt.Errorf("update crawl state: %w", err)
	}

	if err := msg.Ack(); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

// parseCrawlID validates event metadata and parses the crawl job UUID.
func parseCrawlID(meta *birdwatcherv1.EventMetadata) (uuid.UUID, error) {
	if meta == nil || meta.JobId == "" {
		return uuid.Nil, errors.New("missing event metadata or job id")
	}

	crawlID, err := uuid.Parse(meta.JobId)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parse job id: %w", err)
	}

	return crawlID, nil
}
