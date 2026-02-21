package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Anacardo89/order_svc_hex/order_api/config"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/in/http/rest/orderorchestrator"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/out/rpc/grpc/orderreader"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/log"
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
	lokiHook := log.NewLokiHook("http://loki:3100", map[string]string{
		"app": "order_api",
		"env": "dev",
	})
	log.Init(lokiHook)
	shutdown, err := observability.InitTracer("order_api")
	if err != nil {
		log.Log.Error("failed to create exporter", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Log.Error("error shutting down tracer", "error", err)
		}
	}()
	orderWriter, err := initMessaging(cfg.Kafka)
	if err != nil {
		log.Log.Error("failed to init orderwriter", "error", err)
		os.Exit(1)
	}
	defer orderWriter.Close()
	orderReader, err := orderreader.NewOrderReaderClient(cfg.GRPC)
	if err != nil {
		log.Log.Error("failed to init orderreader", "error", err)
		os.Exit(1)
	}
	defer orderReader.Close()
	orderHandler := orderorchestrator.NewOrderHandler(orderReader, orderWriter)
	orderServer := orderorchestrator.NewServer(&cfg.Server, orderHandler)

	stopChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Execution
	go func() {
		log.Log.Info("server listening on", "address", cfg.Server.Port)
		errChan <- orderServer.Start()
	}()

	// Shutdown
	select {
	case sig := <-stopChan:
		log.Log.Info("Shutting down server", "signal", sig)
		orderServer.Shutdown()
		log.Log.Info("Server stopped gracefully")
	case err := <-errChan:
		log.Log.Error("server error", "error", err)
	}
}
