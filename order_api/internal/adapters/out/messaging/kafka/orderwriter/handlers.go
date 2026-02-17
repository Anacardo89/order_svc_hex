package orderwriter

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

func (c *OrderWriterClient) PublishCreate(ctx context.Context, req *core.CreateOrder) error {
	event := OrderCreatedEvent{
		Items:  req.Items,
		Status: string(req.Status),
	}
	return c.producerCreated.publish(ctx, "", event)
}

func (c *OrderWriterClient) PublishStatusUpdate(ctx context.Context, req *core.UpdateOrderStatus) error {
	event := OrderStatusUpdatedEvent{
		ID:     req.ID,
		Status: string(req.Status),
	}
	return c.producerStatusUpdate.publish(ctx, req.ID, event)
}
