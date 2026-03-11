package main

import (
	"fmt"

	"github.com/Anacardo89/order_svc_hex/order_api/config"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/adapters/out/messaging/kafka/orderwriter"
	"github.com/Anacardo89/order_svc_hex/order_api/pkg/events"
	"go.opentelemetry.io/otel/metric"
)

func initMessaging(cfg config.Kafka, meter metric.Meter, m *orderwriter.ProducerMetrics) (*orderwriter.OrderWriterClient, error) {
	conn := events.NewKafkaConnection(cfg.Brokers)
	orderWriterClient, err := orderwriter.NewOrderWriterClient(conn, cfg.Topics, meter, m)
	if err != nil {
		return nil, fmt.Errorf("failed to create Order Writer: %s", err)
	}
	return orderWriterClient, nil
}
