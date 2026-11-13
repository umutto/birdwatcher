package nats

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"uuid"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
)

// NewEventMetadata constructs standard EventMetadata with a deterministic event ID and timestamp.
func NewEventMetadata(jobID string, keys ...string) *birdwatcherv1.EventMetadata {
	sum := sha256.Sum256([]byte(jobID + ":" + strings.Join(keys, ":")))
	return &birdwatcherv1.EventMetadata{
		EventId:   uuid.UUID(sum[:16]).String(),
		JobId:     jobID,
		Timestamp: timestamppb.Now(),
	}
}

// metadataCarrier is implemented by all Birdwatcher v1 Protobuf event messages.
type metadataCarrier interface {
	GetMeta() *birdwatcherv1.EventMetadata
}

// PublishProto serializes and publishes a Protobuf message to a JetStream subject,
// returning the server publication acknowledgment.
func (c *Client) PublishProto(ctx context.Context, subject string, msg proto.Message) (*jetstream.PubAck, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("nats: context cancelled: %w", err)
	}

	if c == nil || c.JS == nil {
		return nil, errors.New("nats: jetstream not initialized")
	}

	if msg == nil {
		return nil, errors.New("nats: cannot publish nil message")
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("nats: marshal proto: %w", err)
	}

	var opts []jetstream.PublishOpt
	if mc, ok := msg.(metadataCarrier); ok {
		if eventID := mc.GetMeta().GetEventId(); eventID != "" {
			opts = append(opts, jetstream.WithMsgID(eventID))
		}
	}

	ack, err := c.JS.Publish(ctx, subject, data, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats: publish to %q: %w", subject, err)
	}

	c.log.DebugContext(ctx, "nats: message published", slog.String("subject", subject))

	return ack, nil
}
