package orderserver

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

type grpcMetrics struct {
	requests metric.Int64Counter
	duration metric.Float64Histogram
	active   metric.Int64UpDownCounter
}

func NewQryMetrics(meter metric.Meter) (*grpcMetrics, error) {
	reqs, err := meter.Int64Counter("order.gRPC.request.total",
		metric.WithDescription("Total number of gRPC requests handled"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC request counter: %w", err)
	}
	dur, err := meter.Float64Histogram("order.gRPC.request.duration",
		metric.WithDescription("Latency of gRPC requests"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC request duration histogram: %w", err)
	}
	active, err := meter.Int64UpDownCounter("order.gRPC.request.active",
		metric.WithDescription("Number of currently executing gRPC requests"),
		metric.WithUnit("{query}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active gRPC request counter: %w", err)
	}
	return &grpcMetrics{
		requests: reqs,
		duration: dur,
		active:   active,
	}, nil
}
