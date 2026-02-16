package orderdlq

import (
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/events"
)

type DlqClient struct {
	producerDlq *Producer
}

func NewDlqClient(kc *events.KafkaConnection, topic string) (ports.OrderDLQ, error) {
	p, err := NewProducer(kc, topic)
	if err != nil {
		return nil, err
	}
	return &DlqClient{
		producerDlq: p,
	}, nil
}

func (c *DlqClient) Close() {
	c.producerDlq.producer.Flush(5000)
	c.producerDlq.producer.Close()
}
