package orderwriter

import (
	"context"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Anacardo89/order_svc_hex/order_api/pkg/events"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/log"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(kc *events.KafkaConnection, topic string) (*Producer, error) {
	p, err := kc.MakeProducer()
	if err != nil {
		return nil, err
	}
	return &Producer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *Producer) publish(ctx context.Context, key string, payload any) error {
	// Observability
	tracer := otel.Tracer("order_api.kafka")
	msgCtx, span := tracer.Start(ctx, "kafka.publish",
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", p.topic),
			attribute.String("messaging.operation", "publish"),
		),
	)
	traceID, spanID := observability.GetTraceSpan(span)
	defer span.End()

	// Execution
	headers := injectTraceHeaders(msgCtx)
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	deliveryChan := make(chan kafka.Event, 1)
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:     []byte(key),
		Value:   value,
		Headers: headers,
	}
	err = p.producer.Produce(msg, deliveryChan)
	if err != nil {
		log.Log.Error("failed to publish message", "trace_id", traceID, "span_id", spanID, "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "publish failed")
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return m.TopicPartition.Error
		}
	}
	return nil
}

func injectTraceHeaders(ctx context.Context) []kafka.Header {
	headers := []kafka.Header{}
	propagator := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)
	for k, v := range carrier {
		headers = append(headers, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}
	return headers
}
