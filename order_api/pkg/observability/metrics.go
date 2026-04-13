package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func InitMetrics(ctx context.Context, serviceName string) (*prometheus.Exporter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	res, _ := resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(exporter),
	)
	otel.SetMeterProvider(mp)
	return exporter, nil
}

// func InitMetrics(ctx context.Context, serviceName, collectorEndpoint string, readerPeriod time.Duration) (func(context.Context) error, error) {
// 	exporter, err := otlpmetricgrpc.New(ctx,
// 		otlpmetricgrpc.WithEndpoint(collectorEndpoint),
// 		otlpmetricgrpc.WithInsecure(),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
// 	}
// 	res, err := resource.Merge(
// 		resource.Default(),
// 		resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceNameKey.String(serviceName),
// 		),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create resource: %w", err)
// 	}
// 	mp := metric.NewMeterProvider(
// 		metric.WithResource(res),
// 		metric.WithReader(metric.NewPeriodicReader(
// 			exporter,
// 			metric.WithInterval(readerPeriod),
// 		)),
// 	)
// 	otel.SetMeterProvider(mp)
// 	return func(shutdownCtx context.Context) error {
// 		return mp.Shutdown(shutdownCtx)
// 	}, nil
// }
