package orderconsumer

import (
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/events"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	topics   []string
}

func NewConsumer(kc *events.KafkaConnection, groupID string, topics []string) (*Consumer, error) {
	c, err := kc.MakeConsumer(groupID, topics)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		consumer: c,
		topics:   topics,
	}, nil
}
