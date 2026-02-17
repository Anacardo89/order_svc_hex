package main

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Anacardo89/order_svc_hex/order_svc/config"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/in/messaging/kafka/orderconsumer"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/out/messaging/kafka/orderdlq"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/out/store/pgx/orderrepo"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/db"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/events"
)

func initDB(cfg config.Config) (core.OrderRepo, func(), error) {
	dbConn, err := db.Connect(cfg.DB)
	if err != nil {
		return nil, nil, err
	}
	migrationPath := filepath.Join(cfg.AppHome, "db", "migrations")
	if err := db.Migrate(cfg.DB.DSN, migrationPath, db.MigrateUp); err != nil {
		return nil, nil, err
	}
	close := func() { dbConn.Close() }
	return orderrepo.NewRepo(dbConn), close, nil
}

func initMessaging(cfg config.Kafka, repo core.OrderRepo) (*orderconsumer.OrderConsumerClient, func(), error) {
	conn := events.NewKafkaConnection(cfg.Brokers)
	dlqTopic, ok := cfg.Topics["OrderDLQ"]
	if !ok {
		return nil, nil, errors.New("no topic for OrderDLQ defined")
	}
	dlqClient, err := orderdlq.NewDlqClient(conn, dlqTopic)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DLQ Client: ", err)
	}
	consumerTopics := []string{}
	createdTopic, ok := cfg.Topics["OrderCreated"]
	if !ok {
		dlqClient.Close()
		return nil, nil, errors.New("no topic for OrderCreated defined")
	} else {
		consumerTopics = append(consumerTopics, createdTopic)
	}
	updatedTopic, ok := cfg.Topics["OrderStatusUpdated"]
	if !ok {
		dlqClient.Close()
		return nil, nil, errors.New("no topic for OrderStatusUpdated defined")
	} else {
		consumerTopics = append(consumerTopics, updatedTopic)
	}
	orderHandler := orderconsumer.NewOrderHandler(repo)
	orderConsumerClient, err := orderconsumer.NewOrderConsumerClient(conn, cfg.GroupID, consumerTopics, orderHandler, dlqClient)
	if err != nil {
		dlqClient.Close()
		return nil, nil, fmt.Errorf("failed to create Order Client: ", err)
	}
	closeDlq := func() { dlqClient.Close() }
	return orderConsumerClient, closeDlq, nil
}
