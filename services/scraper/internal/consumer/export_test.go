package consumer

import (
	"context"

	"github.com/nats-io/nats.go/jetstream"
)

// HandleMessage exports the internal handleMessage dispatcher for external test packages.
func (c *Consumer) HandleMessage(ctx context.Context, msg jetstream.Msg) {
	c.handleMessage(ctx, msg)
}
