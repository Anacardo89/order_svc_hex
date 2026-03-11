package orderwriter

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type ProducerMetrics struct {
	published metric.Int64Counter
	duration  metric.Float64Histogram
	failed    metric.Int64Counter
}

func NewProducerMetrics(meter metric.Meter) (*ProducerMetrics, error) {
	pub, err := meter.Int64Counter("order.command.published.total",
		metric.WithDescription("Total number of commands published to Kafka"),
		metric.WithUnit("{command}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publish counter: %w", err)
	}
	dur, err := meter.Float64Histogram("order.command.publish.duration",
		metric.WithDescription("Latency of Kafka publish operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create publish duration histogram: %w", err)
	}
	fail, err := meter.Int64Counter("order.command.published.failed",
		metric.WithDescription("Total number of publish failures"),
		metric.WithUnit("{command}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create fail counter: %w", err)
	}
	return &ProducerMetrics{
		published: pub,
		duration:  dur,
		failed:    fail,
	}, nil
}
