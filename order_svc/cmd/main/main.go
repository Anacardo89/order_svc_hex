package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Anacardo89/order_svc_hex/order_svc/config"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/in/messaging/kafka/orderconsumer"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/in/rpc/grpc/orderserver"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/observability"
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
		"service": "order_svc",
	})
	tracerShutdown, err := observability.InitTracer(ctx, "order_svc", cfg.Trace.Endpoint)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to create tracer", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer func() {
		if err := tracerShutdown(ctx); err != nil {
			logger.BaseLogger.Error(ctx, "error shutting down tracer", ports.Field{Key: "error", Value: err})
		}
	}()
	metricsShutdown, err := observability.InitMetrics(ctx, "order_svc", cfg.Metric.Endpoint, cfg.Metric.ReaderPeriod)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to create metrics", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer func() {
		if err := metricsShutdown(ctx); err != nil {
			logger.BaseLogger.Error(ctx, "error shutting down metrics", ports.Field{Key: "error", Value: err})
		}
	}()
	consumerMeter := otel.GetMeterProvider().Meter("order_svc.consumer")
	consumerMetrics, err := orderconsumer.NewConsumerMetrics(consumerMeter)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init producer metrics", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	dbRepo, err := initDB(*cfg)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init db", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer dbRepo.Close()
	orderConsumer, closeDlq, err := initMessaging(cfg.Kafka, dbRepo, consumerMetrics)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to init messaging", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer orderConsumer.Close()
	defer closeDlq()
	gRPCservice := orderserver.NewOrderGRPCService(dbRepo)
	gRPCServer, err := orderserver.NewOrderGRPCServer(cfg.Server.Port, gRPCservice)
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to create gRPC server", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	errSrvChan := make(chan error, 1)
	errEventChan := make(chan error, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Execution
	go func() {
		logger.BaseLogger.Info(ctx, "gRPC server listening on", ports.Field{Key: "address", Value: gRPCServer.Listener.Addr()})
		errSrvChan <- gRPCServer.Server.Serve(gRPCServer.Listener)
	}()
	go func() {
		logger.BaseLogger.Info(ctx, "consumer starting")
		errEventChan <- orderConsumer.Consume(ctx)
	}()

	// Shutdown
	select {
	case sig := <-stopChan:
		logger.BaseLogger.Info(ctx, "Shutting down gRPC server", ports.Field{Key: "signal", Value: sig})
		gRPCServer.Server.GracefulStop()
		logger.BaseLogger.Info(ctx, "Server stopped gracefully")
	case err := <-errSrvChan:
		logger.BaseLogger.Error(ctx, "gRPC server error", ports.Field{Key: "error", Value: err})
	case err := <-errEventChan:
		logger.BaseLogger.Error(ctx, "consumer error", ports.Field{Key: "error", Value: err})
	}
}
