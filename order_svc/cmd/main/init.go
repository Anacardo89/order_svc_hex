package main

import (
	"fmt"
	"path/filepath"

	"github.com/Anacardo89/order_svc_hex/order_svc/config"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/messaging"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/repo"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
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
	return repo.NewOrderRepo(dbConn), close, nil
}

func initEvents(cfg config.Kafka) (ports.OrderConsumer, func(), ports.OrderDLQProducer, func(), *messaging.OrderEventHandler, error) {
	conn := events.NewKafkaConnection(cfg.Brokers)
	allTopics := []string{}
	consumerTopics := []string{}
	dlqTopics := make(map[string]*string)
	rawDLQTopic := ""
	for k, v := range cfg.Topics {
		if v.Name == "" {
			allTopics = append(allTopics, v.DLQ)
			rawDLQTopic = v.DLQ
			continue
		}
		allTopics = append(allTopics, v.Name)
		allTopics = append(allTopics, v.DLQ)
		consumerTopics = append(consumerTopics, v.Name)
		dlqTopics[k] = &v.DLQ
	}
	err := events.EnsureTopics(cfg.Brokers, allTopics, 1)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	consumer, err := conn.MakeConsumer(cfg.GroupID, consumerTopics)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	consumer.Poll(100)
	assigned, err := consumer.Assignment()
	fmt.Println("Assigned partitions:", assigned)
	rawProducer, err := conn.MakeProducer()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	dlqProducer, err := conn.MakeProducer()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	rawDLQProducer := messaging.NewRawDLQProducer(rawProducer, rawDLQTopic)
	orderDLQProducer := messaging.NewOrderDLQProducer(dlqProducer, dlqTopics)
	orderConsumer := messaging.NewOrderConsumer(consumer, rawDLQProducer)
	consumerHandler := messaging.NewOrderEventHandler(cfg.QueueSize)
	closeConsumer := func() { rawProducer.Close(); consumer.Close() }
	closeDLQProducer := func() { dlqProducer.Close() }
	return orderConsumer, closeConsumer, orderDLQProducer, closeDLQProducer, consumerHandler, nil
}
