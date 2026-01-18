package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Anacardo89/order_svc_hex/config"
	"github.com/Anacardo89/order_svc_hex/internal/adapters/api/grpc"
	kafkaevents "github.com/Anacardo89/order_svc_hex/internal/adapters/events/kafka"
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
	dbRepo, closeDB, err := initDB(*cfg)
	if err != nil {
		slog.Error("failed to init db", "error", err)
		os.Exit(1)
	}
	defer closeDB()
	eventConsumer, closeConsumer, dlqProducer, closeDLQProducer, consumerHandler, err := initEvents(cfg.Kafka)
	if err != nil {
		slog.Error("failed to init Kafka", "error", err)
		os.Exit(1)
	}
	defer closeConsumer()
	defer closeDLQProducer()
	workerPool := kafkaevents.NewWorkerPool(
		dbRepo,
		dlqProducer,
		consumerHandler,
		cfg.Kafka.BatchSize,
		cfg.Kafka.BatchTimeout,
		cfg.Kafka.WorkerPoolSize,
	)
	gRPCservice := grpc.NewOrderGRPCService(dbRepo)
	gRPCServer, err := grpc.NewOrderGRPCServer(cfg.Server.Port, gRPCservice)
	if err != nil {
		slog.Error("failed to create gRPC server", "error", err)
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	errSrvChan := make(chan error, 1)
	errEventChan := make(chan error, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	workerPool.Start(ctx)
	go func() {
		slog.Info("gRPC server listening on", "address", gRPCServer.Listener.Addr())
		errSrvChan <- gRPCServer.Server.Serve(gRPCServer.Listener)
	}()
	go func() {
		slog.Info("consumer starting")
		errEventChan <- eventConsumer.Consume(ctx, consumerHandler)
	}()

	select {
	case sig := <-stopChan:
		slog.Info("Shutting down gRPC server", "signal", sig)
		gRPCServer.Server.GracefulStop()
		slog.Info("Server stopped gracefully")
	case err := <-errSrvChan:
		slog.Error("gRPC server error", "error", err)
	case err := <-errEventChan:
		slog.Error("consumer error", "error", err)
	}
}
