package events

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConnection struct {
	Brokers string
}

func NewKafkaConnection(brokers string) *KafkaConnection {
	return &KafkaConnection{
		Brokers: brokers,
	}
}

func (c *KafkaConnection) MakeConsumer(groupID string, topics []string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	err = consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topics: %w", err)
	}
	return consumer, nil
}

func (c *KafkaConnection) MakeProducer() (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
	})
	if err != nil {
		return nil, err
	}
	return producer, nil
}
