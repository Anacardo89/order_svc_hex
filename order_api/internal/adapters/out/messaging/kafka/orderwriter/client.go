package orderwriter

import (
	"fmt"
	"sync"

	"github.com/Anacardo89/order_svc_hex/order_api/pkg/events"
)

type OrderWriterClient struct {
	producerCreated      *Producer
	producerStatusUpdate *Producer
}

func NewOrderWriterClient(kc *events.KafkaConnection, topics map[string]string) (*OrderWriterClient, error) {
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
	return &OrderWriterClient{
		producerCreated:      pc,
		producerStatusUpdate: ps,
	}, nil
}

func (c *OrderWriterClient) Close() {
	var wg sync.WaitGroup
	wg.Go(func() {
		c.producerCreated.producer.Flush(5000)
		c.producerCreated.producer.Close()
	})
	wg.Go(func() {
		c.producerStatusUpdate.producer.Flush(5000)
		c.producerStatusUpdate.producer.Close()
	})
	wg.Wait()
}
