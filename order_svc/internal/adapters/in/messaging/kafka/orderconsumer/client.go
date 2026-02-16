package orderconsumer

import (
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/events"
)

type OrderConsumerClient struct {
	orderConsumer  *Consumer
	handler        ports.OrderConsumer
	dlqClient      ports.OrderDLQ
	workerPoolSize int
}

func NewOrderConsumerClient(
	kc *events.KafkaConnection,
	groupId string,
	topics []string,
	handler ports.OrderConsumer,
	dlqClient ports.OrderDLQ,
) (*OrderConsumerClient, error) {
	c, err := NewConsumer(kc, groupId, topics)
	if err != nil {
		return nil, err
	}
	return &OrderConsumerClient{
		orderConsumer: c,
		handler:       handler,
		dlqClient:     dlqClient,
	}, nil
}

func (c *OrderConsumerClient) Close() {
	c.orderConsumer.consumer.Close()
}
