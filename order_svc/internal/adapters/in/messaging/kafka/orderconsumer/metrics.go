package orderconsumer

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type ConsumerMetrics struct {
	consumed    metric.Int64Counter
	duration    metric.Float64Histogram
	failed      metric.Int64Counter
	dlqProduced metric.Int64Counter
	dataLoss    metric.Int64Counter
}

func NewConsumerMetrics(meter metric.Meter) (*ConsumerMetrics, error) {
	cons, err := meter.Int64Counter("order.event.consumed.total",
		metric.WithDescription("Total number of events consumed from Kafka"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create events consumed counter: %w", err)
	}
	dur, err := meter.Float64Histogram("order.event.processing.duration",
		metric.WithDescription("Time taken to process a Kafka event"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create events processing duration histogram: %w", err)
	}
	fail, err := meter.Int64Counter("order.event.consumed.failed",
		metric.WithDescription("Total number of event processing failures"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create events failed counter: %w", err)
	}
	dlqCounter, err := meter.Int64Counter("order.event.dlq.total",
		metric.WithDescription("Total number of DLQ publish"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ counter: %w", err)
	}
	dlqFail, err := meter.Int64Counter("order.event.dlq.fail",
		metric.WithDescription("Total number of DLQ publish faillures"),
		metric.WithUnit("{event}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ fail counter: %w", err)
	}
	return &ConsumerMetrics{
		consumed:    cons,
		duration:    dur,
		failed:      fail,
		dlqProduced: dlqCounter,
		dataLoss:    dlqFail,
	}, nil
}
