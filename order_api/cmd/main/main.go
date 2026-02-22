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
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/out/rpc/grpc/orderreader"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/observability"
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
	logger.BaseLogger = logger.NewLogger("http://loki:3100/loki/api/v1/push", map[string]string{
		"service": "order_api",
	})
	shutdown, err := observability.InitTracer("order_api")
	if err != nil {
		logger.BaseLogger.Error(ctx, "failed to create exporter", ports.Field{Key: "error", Value: err})
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			logger.BaseLogger.Error(ctx, "error shutting down tracer", ports.Field{Key: "error", Value: err})
		}
	}()
	orderWriter, err := initMessaging(cfg.Kafka)
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
	orderServer := orderorchestrator.NewServer(&cfg.Server, orderHandler)

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
