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

func initDB(cfg config.Config) (*orderrepo.OrderRepo, error) {
	dbConn, err := db.Connect(cfg.DB)
	if err != nil {
		return nil, err
	}
	migrationPath := filepath.Join(cfg.AppHome, "db", "migrations")
	if err := db.Migrate(cfg.DB.DSN, migrationPath, db.MigrateUp); err != nil {
		dbConn.Close()
		return nil, err
	}
	return orderrepo.NewRepo(dbConn), nil
}

func initMessaging(cfg config.Kafka, repo core.OrderRepo) (*orderconsumer.OrderConsumerClient, func(), error) {
	conn := events.NewKafkaConnection(cfg.Brokers)
	allTopics := []string{}
	for _, v := range cfg.Topics {
		allTopics = append(allTopics, v)
	}
	if err := events.EnsureTopics(cfg.Brokers, allTopics, 1); err != nil {
		return nil, nil, fmt.Errorf("failed to ensure topics: %s", err)
	}
	dlqTopic, ok := cfg.Topics["OrderDLQ"]
	if !ok {
		return nil, nil, errors.New("no topic for OrderDLQ defined")
	}
	dlqClient, err := orderdlq.NewDlqClient(conn, dlqTopic)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DLQ Client: %s", err)
	}
	closeDlq := func() { dlqClient.Close() }
	consumerTopics := []string{}
	createdTopic, ok := cfg.Topics["OrderCreated"]
	if !ok {
		closeDlq()
		return nil, nil, errors.New("no topic for OrderCreated defined")
	} else {
		consumerTopics = append(consumerTopics, createdTopic)
	}
	updatedTopic, ok := cfg.Topics["OrderStatusUpdated"]
	if !ok {
		closeDlq()
		return nil, nil, errors.New("no topic for OrderStatusUpdated defined")
	} else {
		consumerTopics = append(consumerTopics, updatedTopic)
	}
	orderHandler := orderconsumer.NewOrderHandler(repo)
	orderConsumerClient, err := orderconsumer.NewOrderConsumerClient(conn, cfg.GroupID, consumerTopics, orderHandler, dlqClient)
	if err != nil {
		closeDlq()
		return nil, nil, fmt.Errorf("failed to create Order Client: %s", err)
	}
	return orderConsumerClient, closeDlq, nil
}
