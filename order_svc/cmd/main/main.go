package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Anacardo89/order_svc_hex/order_svc/config"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/in/rpc/grpc/orderserver"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/log"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/observability"
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
		"app": "order_svc",
		"env": "dev",
	})
	log.Init(lokiHook)
	shutdown, err := observability.InitTracer("order_svc")
	if err != nil {
		log.Log.Error("failed to create exporter", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Log.Error("error shutting down tracer", "error", err)
		}
	}()
	dbRepo, err := initDB(*cfg)
	if err != nil {
		log.Log.Error("failed to init db", "error", err)
		os.Exit(1)
	}
	defer dbRepo.Close()
	orderConsumer, closeDlq, err := initMessaging(cfg.Kafka, dbRepo)
	if err != nil {
		log.Log.Error("failed to init messaging", "error", err)
		os.Exit(1)
	}
	defer orderConsumer.Close()
	defer closeDlq()
	gRPCservice := orderserver.NewOrderGRPCService(dbRepo)
	gRPCServer, err := orderserver.NewOrderGRPCServer(cfg.Server.Port, gRPCservice)
	if err != nil {
		log.Log.Error("failed to create gRPC server", "error", err)
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	errSrvChan := make(chan error, 1)
	errEventChan := make(chan error, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// Execution
	go func() {
		log.Log.Info("gRPC server listening on", "address", gRPCServer.Listener.Addr())
		errSrvChan <- gRPCServer.Server.Serve(gRPCServer.Listener)
	}()
	go func() {
		log.Log.Info("consumer starting")
		errEventChan <- orderConsumer.Consume(ctx)
	}()

	// Shutdown
	select {
	case sig := <-stopChan:
		log.Log.Info("Shutting down gRPC server", "signal", sig)
		gRPCServer.Server.GracefulStop()
		log.Log.Info("Server stopped gracefully")
	case err := <-errSrvChan:
		log.Log.Error("gRPC server error", "error", err)
	case err := <-errEventChan:
		log.Log.Error("consumer error", "error", err)
	}
}
