package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Anacardo89/order_svc_hex/order_api/config"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/in/http/rest/orderorchestrator"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/out/messaging/kafka/orderwriter"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/out/rpc/grpc/orderreader"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
	"go.opentelemetry.io/otel"
)

func main() {
	// Setup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	logger.BaseLogger = logger.NewLogger(cfg.Log.Endpoint, map[string]string{
		"service": "order_api",
	})
	tracerShutdown, err := observability.InitTracer(ctx, "order_api", cfg.Trace.Endpoint)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to create tracer", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer func() {
		if err := tracerShutdown(ctx); err != nil {
			logger.BaseLogger.Error(ctx, "error shutting down tracer", ports.Field{Key: "error", Value: err})
		}
	}()
	metricsShutdown, err := observability.InitMetrics(ctx, "order_api", cfg.Metric.Endpoint, cfg.Metric.ReaderPeriod)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to create metrics", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer func() {
		if err := metricsShutdown(ctx); err != nil {
			logger.BaseLogger.Error(ctx, "error shutting down metrics", ports.Field{Key: "error", Value: err})
		}
	}()
	restMeter := otel.GetMeterProvider().Meter("order_api.rest")
	producerMeter := otel.GetMeterProvider().Meter("order_api.producer")
	restMetrics, err := orderorchestrator.NewReqMetrics(restMeter)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init rest metrics", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	producerMetrics, err := orderwriter.NewProducerMetrics(producerMeter)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init producer metrics", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	orderWriter, err := initMessaging(cfg.Kafka, producerMeter, producerMetrics)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init orderwriter", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer orderWriter.Close()
	var ow ports.OrderWriter = orderWriter
	orderReader, err := orderreader.NewOrderReaderClient(cfg.GRPC)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init orderreader", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer orderReader.Close()
	var or ports.OrderReader = orderReader
	orderHandler := orderorchestrator.NewOrderHandler(or, ow)
	orderServer := orderorchestrator.NewServer(&cfg.Server, orderHandler, restMetrics)

	stopChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Execution
	go func() {
		logger.BaseLogger.Info(ctx, "Starting server on", ports.Field{Key: "port", Value: cfg.Server.Port})
		errChan <- orderServer.Start()
	}()

	// Shutdown
	select {
	case sig := <-stopChan:
		logger.BaseLogger.Info(ctx, "Shutting down server", ports.Field{Key: "signal", Value: sig})
		orderServer.Shutdown()
		logger.BaseLogger.Info(ctx, "Server stopped gracefully")
	case err := <-errChan:
		logger.BaseLogger.Error(ctx, "server error", ports.Field{Key: "error", Value: err})
	}
}
