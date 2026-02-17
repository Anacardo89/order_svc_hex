package orderwriter

import (
	"context"
	"encoding/json"

	"github.com/Anacardo89/order_svc_hex/order_api/pkg/events"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(kc *events.KafkaConnection, topic string) (*Producer, error) {
	p, err := kc.MakeProducer()
	if err != nil {
		return nil, err
	}
	return &Producer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *Producer) publish(ctx context.Context, key string, payload any) error {
	value, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	deliveryChan := make(chan kafka.Event, 1)
	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(key),
		Value: value,
	}, deliveryChan)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			return m.TopicPartition.Error
		}
	}
	return nil
}
