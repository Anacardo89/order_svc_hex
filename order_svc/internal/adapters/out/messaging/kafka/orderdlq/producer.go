package orderdlq

import (
	"context"
	"encoding/json"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/events"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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

func (p *Producer) publish(ctx context.Context, key string, payload any, reason string) error {
	// Observability
	tracer := otel.Tracer("order_svc.kafka.dlq")
	ctx, span := tracer.Start(ctx, "kafka.publish",
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", p.topic),
			attribute.String("messaging.operation", "publish"),
			attribute.String("error.reason", reason),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	deliveryChan := make(chan kafka.Event, 1)
	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}, deliveryChan)
	if err != nil {
		log.Error(ctx, "publish dlq failed", ports.Field{Key: "error", Value: err})
		span.RecordError(err)
		span.SetStatus(codes.Error, "publish dlq failed")
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
