package orderwriter

import (
	"fmt"

	"github.com/Anacardo89/order_svc_hex/order_api/pkg/events"
)

type KafkaClient struct {
	producerCreated      *Producer
	producerStatusUpdate *Producer
}

func NewKafkaClient(kc *events.KafkaConnection, topics map[string]string) (*KafkaClient, error) {
	createdTopic, ok := topics[string(TopicOrderCreated)]
	if !ok {
		return nil, fmt.Errorf("missing topic: %s", TopicOrderCreated)
	}
	pc, err := NewProducer(kc, createdTopic)
	if err != nil {
		return nil, err
	}
	updatedTopic, ok := topics[string(TopicOrderStatusUpdated)]
	if !ok {
		return nil, fmt.Errorf("missing topic: %s", TopicOrderStatusUpdated)
	}
	ps, err := NewProducer(kc, updatedTopic)
	if err != nil {
		return nil, err
	}
	return &KafkaClient{
		producerCreated:      pc,
		producerStatusUpdate: ps,
	}, nil
}

func (c *KafkaClient) Close() {
	c.producerCreated.producer.Flush(5000)
	c.producerCreated.producer.Close()
	c.producerStatusUpdate.producer.Flush(5000)
	c.producerStatusUpdate.producer.Close()
}
