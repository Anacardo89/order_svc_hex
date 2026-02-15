package orderwriter

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

func (c *KafkaClient) PublishCreate(ctx context.Context, req *core.CreateOrderReq) error {
	event := OrderCreatedEvent{
		Items:  req.Items,
		Status: string(req.Status),
	}
	return c.producerCreated.publish(ctx, "", event)
}

func (c *KafkaClient) PublishStatusUpdate(ctx context.Context, req *core.UpdateOrderStatusReq) error {
	event := OrderStatusUpdatedEvent{
		ID:     req.ID,
		Status: string(req.Status),
	}
	return c.producerStatusUpdate.publish(ctx, req.ID, event)
}
