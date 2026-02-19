package orderconsumer

import (
	"context"
	"errors"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func (c *OrderConsumerClient) Consume(ctx context.Context) error {
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
			c.handleMessage(msg)
		}
	}
}

func (c *OrderConsumerClient) handleMessage(msg *kafka.Message) {
	// Error Hanfling
	fail := func(ctx context.Context, span trace.Span, msg *kafka.Message, reason string, err error) {
		dlqMsg := makeDlqMessage(msg, reason, err)
		if err := c.dlqClient.PublishDLQ(ctx, dlqMsg); err != nil {
			slog.Error("failed to send to DLQ on error", "error", err, "dlq_msg", dlqMsg)
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, reason)
	}

	// Observability
	msgCtx := extractContextFromKafka(msg)
	tracer := otel.Tracer("order_svc.kafka")
	msgCtx, span := tracer.Start(msgCtx, "kafka.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.operation", "consume"),
			attribute.String("messaging.destination", *msg.TopicPartition.Topic),
		),
	)
	defer span.End()

	// Execution
	order, err := mapEventPaylodToOrder(msg)
	if err != nil {
		slog.Error("failed to unmarshal payload", "error", err)
		fail(msgCtx, span, msg, "unmarshal_failed", err)
		return
	}
	switch *msg.TopicPartition.Topic {
	case "orders.created":
		if err := c.handler.OnOrderCreated(msgCtx, *order); err != nil {
			slog.Error("failed to handle order created", "error", err)
			fail(msgCtx, span, msg, "handler_error", err)
		}
	case "orders.status_updated":
		if err := c.handler.OnOrderStatusUpdated(msgCtx, *order); err != nil {
			slog.Error("failed to handle order status updated", "error", err)
			fail(msgCtx, span, msg, "handler_error", err)
		}
	default:
		slog.Error("message with unknown topic", "msg", msg)
		fail(msgCtx, span, msg, "unknown_topic", errors.New("unknown topic"))
	}
	if _, err := c.orderConsumer.consumer.CommitMessage(msg); err != nil {
		slog.Error("failed to commit offset", "error", err)
		fail(msgCtx, span, msg, "offset_commit_failed", err)
	}
}

func extractContextFromKafka(msg *kafka.Message) context.Context {
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier{}
	for _, h := range msg.Headers {
		carrier[h.Key] = string(h.Value)
	}
	return propagator.Extract(context.Background(), carrier)
}
