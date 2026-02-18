package orderconsumer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func (c *OrderConsumerClient) Consume(ctx context.Context) error {
	// Error handling
	fail := func(msg *kafka.Message, reason string, err error) {
		dlqMsg := makeDlqMessage(msg, reason, err)
		if err := c.dlqClient.PublishDLQ(ctx, dlqMsg); err != nil {
			slog.Error("failed to send to DLQ on error", "error", err, "dlq_msg", dlqMsg)
		}
	}

	// Execution
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.orderConsumer.consumer.ReadMessage(-1)
			if err != nil {
				slog.Error("failed to read message", "error", err)
				continue
			}
			order, err := mapEventPaylodToOrder(msg)
			if err != nil {
				slog.Error("failed to unmarshal payload", "error", err)
				fail(msg, "unmarshal_failed", err)
				continue
			}
			switch *msg.TopicPartition.Topic {
			case "orders.created":
				if err := c.handler.OnOrderCreated(ctx, *order); err != nil {
					slog.Error("failed to handle order created", "error", err)
					fail(msg, "handler_error", err)
				}
			case "orders.status_updated":
				if err := c.handler.OnOrderStatusUpdated(ctx, *order); err != nil {
					slog.Error("failed to handle order status updated", "error", err)
					fail(msg, "handler_error", err)
				}
			default:
				slog.Error("message with unknown topic", "msg", msg)
				fail(msg, "unknown_topic", errors.New("unknown topic"))
			}
			if _, err := c.orderConsumer.consumer.CommitMessage(msg); err != nil {
				slog.Error("failed to commit offset", "error", err)
			}
		}
	}
}
