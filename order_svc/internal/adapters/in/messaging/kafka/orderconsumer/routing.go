package orderconsumer

import (
	"context"
	"errors"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/observability"
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
				logger.BaseLogger.Error(ctx, "failed to read message", ports.Field{Key: "error", Value: err})
				continue
			}
			c.handleMessage(msg)
		}
	}
}

func (c *OrderConsumerClient) handleMessage(msg *kafka.Message) {
	// Error Handling
	fail := func(ctx context.Context, span trace.Span, msg *kafka.Message, reason string, err error) {
		span.RecordError(err)
		span.SetStatus(codes.Error, reason)
		traceID, spanID := observability.GetTraceSpan(span)
		dlqMsg := makeDlqMessage(msg, traceID, spanID, reason, err)
		if err := c.dlqClient.PublishDLQ(ctx, dlqMsg); err != nil {
			log := logger.LogFromSpan(span, logger.BaseLogger)
			log.Error(ctx, "failed to send to DLQ on error", ports.Field{Key: "error", Value: err}, ports.Field{Key: "dlq_msg", Value: dlqMsg})
		}
	}

	// Observability
	msgCtx := extractContextFromKafka(msg)
	tracer := otel.Tracer("order_svc.kafka")
	msgCtx, span := tracer.Start(msgCtx, "kafka.consume",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.operation", "consume"),
			attribute.String("messaging.source", *msg.TopicPartition.Topic),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
	order, err := mapEventPaylodToOrder(msg)
	if err != nil {
		log.Error(msgCtx, "failed to unmarshal payload", ports.Field{Key: "error", Value: err})
		fail(msgCtx, span, msg, "unmarshal_failed", err)
		return
	}
	switch *msg.TopicPartition.Topic {
	case "orders.created":
		if err := c.handler.OnOrderCreated(msgCtx, *order); err != nil {
			log.Error(msgCtx, "failed to handle order created", ports.Field{Key: "error", Value: err})
			fail(msgCtx, span, msg, "handler_error", err)
		}
	case "orders.status_updated":
		if err := c.handler.OnOrderStatusUpdated(msgCtx, *order); err != nil {
			log.Error(msgCtx, "failed to handle order status updated", ports.Field{Key: "error", Value: err})
			fail(msgCtx, span, msg, "handler_error", err)
		}
	default:
		log.Error(msgCtx, "message with unknown topic", ports.Field{Key: "msg", Value: msg})
		fail(msgCtx, span, msg, "unknown_topic", errors.New("unknown topic"))
	}
	if _, err := c.orderConsumer.consumer.CommitMessage(msg); err != nil {
		log.Error(msgCtx, "failed to commit offset", ports.Field{Key: "error", Value: err})
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
