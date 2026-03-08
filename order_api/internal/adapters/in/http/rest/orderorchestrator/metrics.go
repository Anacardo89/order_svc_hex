package orderorchestrator

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type ReqMetrics struct {
	counter  metric.Int64Counter
	duration metric.Float64Histogram
	active   metric.Int64UpDownCounter
}

func NewReqMetrics(meter metric.Meter) (*ReqMetrics, error) {
	counter, err := meter.Int64Counter("http.server.request.total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request counter: %w", err)
	}
	duration, err := meter.Float64Histogram("http.server.request.duration",
		metric.WithDescription("Time taken to process HTTP requests"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request duration histogram: %w", err)
	}
	active, err := meter.Int64UpDownCounter("http.server.active_requests",
		metric.WithDescription("Number of active HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active requests counter: %w", err)
	}
	return &ReqMetrics{
		counter:  counter,
		duration: duration,
		active:   active,
	}, nil
}
