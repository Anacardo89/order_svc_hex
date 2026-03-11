package orderconsumer

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type ConsumerMetrics struct {
	consumed metric.Int64Counter
	duration metric.Float64Histogram
	failed   metric.Int64Counter
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
	return &ConsumerMetrics{
		consumed: cons,
		duration: dur,
		failed:   fail,
	}, nil
}
