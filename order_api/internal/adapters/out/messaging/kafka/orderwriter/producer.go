package orderwriter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/events"
)

type Producer struct {
	producer     *kafka.Producer
	metrics      *ProducerMetrics
	topic        string
	registration metric.Registration
}

func NewProducer(kc *events.KafkaConnection, topic string, meter metric.Meter, m *ProducerMetrics) (*Producer, error) {
	p, err := kc.MakeProducer()
	if err != nil {
		return nil, err
	}
	gauge, err := meter.Int64ObservableGauge("order.producer.queue.depth",
		metric.WithDescription("Current messages in queue"),
	)
	if err != nil {
		return nil, err
	}
	producer := &Producer{
		producer: p,
		metrics:  m,
		topic:    topic,
	}
	reg, err := meter.RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
		obs.ObserveInt64(gauge, int64(p.Len()), metric.WithAttributes(
			attribute.String("messaging.destination", producer.topic),
		))
		return nil
	}, gauge)
	if err != nil {
		return nil, fmt.Errorf("failed to register gauge callback: %w", err)
	}
	producer.registration = reg
	return producer, nil
}

func (p *Producer) publish(ctx context.Context, key string, payload any) error {
	// Error handling
	fail := func(msg string, span trace.Span, metricAttrs metric.MeasurementOption, err error) {
		p.metrics.failed.Add(ctx, 1, metricAttrs)
		span.RecordError(err)
		span.SetStatus(codes.Error, msg)
	}

	// Observability
	start := time.Now()
	metricAttrs := metric.WithAttributes(
		attribute.String("messaging.destination", p.topic),
	)
	tracer := otel.Tracer("order_api.kafka")
	msgCtx, span := tracer.Start(ctx, "kafka.publish",
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination", p.topic),
			attribute.String("messaging.operation", "publish"),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
	headers := injectTraceHeaders(msgCtx)
	value, err := json.Marshal(payload)
	if err != nil {
		log.Error(msgCtx, "failed to marshal message", ports.Field{Key: "error", Value: err})
		fail("marshal failed", span, metricAttrs, err)
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
		log.Error(msgCtx, "failed to publish message", ports.Field{Key: "error", Value: err})
		fail("publish failed", span, metricAttrs, err)
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		elapsed := time.Since(start).Seconds()
		p.metrics.duration.Record(ctx, elapsed, metricAttrs)
		if m.TopicPartition.Error != nil {
			log.Error(msgCtx, "failed to ack publish", ports.Field{Key: "error", Value: m.TopicPartition.Error})
			fail("publish ack failed", span, metricAttrs, m.TopicPartition.Error)
			return m.TopicPartition.Error
		}
		p.metrics.published.Add(ctx, 1, metricAttrs)
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
