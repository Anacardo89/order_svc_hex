package events

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Connection struct {
	Brokers string
}

func NewKafkaConnection(brokers []string) *Connection {
	return &Connection{
		Brokers: strings.Join(brokers, ","),
	}
}

func (c *Connection) GetConsumer(groupID string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func (c *Connection) GetProducer() (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
	})
	if err != nil {
		return nil, err
	}
	return producer, nil
}
