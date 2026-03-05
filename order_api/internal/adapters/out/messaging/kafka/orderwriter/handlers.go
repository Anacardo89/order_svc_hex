package orderwriter

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

func (c *OrderWriterClient) PublishCreate(ctx context.Context, cmd *core.CreateOrderCmd) error {
	event := OrderCreatedEvent{
		Items:  cmd.Items,
		Status: string(cmd.Status),
	}
	return c.producerCreated.publish(ctx, "", event)
}

func (c *OrderWriterClient) PublishStatusUpdate(ctx context.Context, cmd *core.UpdateOrderStatusCmd) error {
	event := OrderStatusUpdatedEvent{
		ID:     cmd.ID,
		Status: string(cmd.Status),
	}
	return c.producerStatusUpdate.publish(ctx, cmd.ID, event)
}
