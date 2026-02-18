package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Anacardo89/order_svc_hex/order_api/config"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/in/http/rest/orderorchestrator"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/out/rpc/grpc/orderreader"
)

func main() {
	// Setup
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	orderWriter, err := initMessaging(cfg.Kafka)
	if err != nil {
		slog.Error("failed to init orderwriter", "error", err)
		os.Exit(1)
	}
	defer orderWriter.Close()
	orderReader, err := orderreader.NewOrderReaderClient(cfg.GRPC.Port)
	if err != nil {
		slog.Error("failed to init orderreader", "error", err)
		os.Exit(1)
	}
	defer orderReader.Close()
	orderHandler := orderorchestrator.NewOrderHandler(orderReader, orderWriter)
	orderServer := orderorchestrator.NewServer(&cfg.Server, orderHandler)

	stopChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server listening on", "address", cfg.Server.Port)
		errChan <- orderServer.Start()
	}()

	select {
	case sig := <-stopChan:
		slog.Info("Shutting down server", "signal", sig)
		orderServer.Shutdown()
		slog.Info("Server stopped gracefully")
	case err := <-errChan:
		slog.Error("server error", "error", err)
	}
}
