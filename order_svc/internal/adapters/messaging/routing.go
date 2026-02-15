package messaging

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
)

func (c *OrderConsumer) Consume(ctx context.Context, handler ports.OrderEventHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				slog.Error("[Consume - ReadMessage]", "error", err)
				continue
			}
			var event core.OrderEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				slog.Error("[Consume] - json_unmarshal_failed", "error", err, "payload", msg.Value)
				if err := c.dlqProducer.PublishRawDLQ(ctx, msg, "json_unmarshal_failed"); err != nil {
					slog.Error("[Consume] - failed to sent to DLQ on json_unmarshal_failed", "error", err, "payload", string(msg.Value))
				}
				continue
			}
			switch event.EventType {
			case core.EventOrderCreated:
				if err := handler.OnOrderCreated(ctx, event); err != nil {
					slog.Error("[OnOrderCreated]", "error", err, "payload", event)
					continue
				}
			case core.EventOrderStatusUpdated:
				if err := handler.OnOrderStatusUpdated(ctx, event); err != nil {
					slog.Error("[OnOrderStatusUpdated]", "error", err, "payload", event)
					continue
				}
			default:
				slog.Error("[Consume] - invalid_eventType", "payload", event)
				if err := c.dlqProducer.PublishRawDLQ(ctx, msg, "invalid_eventType"); err != nil {
					slog.Error("[Consume] - failed to sent to DLQ on invalid_eventType", "error", err, "payload", string(msg.Value))
				}
			}
		}
	}
}
