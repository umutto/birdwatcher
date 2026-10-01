package nats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

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

	ack, err := c.JS.Publish(ctx, subject, data)
	if err != nil {
		return nil, fmt.Errorf("nats: publish to %q: %w", subject, err)
	}

	c.log.DebugContext(ctx, "nats: message published", slog.String("subject", subject))

	return ack, nil
}
