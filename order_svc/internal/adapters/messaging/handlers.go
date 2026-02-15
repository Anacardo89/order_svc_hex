package messaging

import (
	"context"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// Generic DLQ
type RawDLQProducer interface {
	PublishRawDLQ(ctx context.Context, msg *kafka.Message, reason string) error
}

type KafkaRawDLQProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewRawDLQProducer(producer *kafka.Producer, topic string) RawDLQProducer {
	return &KafkaRawDLQProducer{
		producer: producer,
		topic:    topic,
	}
}

// Event-specific DLQ
type OrderDLQProducer struct {
	producer *kafka.Producer
	topics   map[string]*string
}

func NewOrderDLQProducer(producer *kafka.Producer, topics map[string]*string) ports.OrderDLQProducer {
	return &OrderDLQProducer{
		producer: producer,
		topics:   topics,
	}
}

// Consumer
type OrderConsumer struct {
	consumer    *kafka.Consumer
	dlqProducer RawDLQProducer
}

func NewOrderConsumer(consumer *kafka.Consumer, dlqProducer RawDLQProducer) ports.OrderConsumer {
	return &OrderConsumer{
		consumer:    consumer,
		dlqProducer: dlqProducer,
	}
}

type OrderEventHandler struct {
	createdQueue       chan core.OrderEvent
	statusUpdatedQueue chan core.OrderEvent
}

func NewOrderEventHandler(queueSize int) *OrderEventHandler {
	return &OrderEventHandler{
		createdQueue:       make(chan core.OrderEvent, queueSize),
		statusUpdatedQueue: make(chan core.OrderEvent, queueSize),
	}
}

// Worker Pool
type WorkerPool struct {
	repo         core.OrderRepo
	dlqProducer  ports.OrderDLQProducer
	handler      *OrderEventHandler
	batchSize    int
	batchTimeout time.Duration
	workers      int
}

func NewWorkerPool(
	repo core.OrderRepo,
	dlqProducer ports.OrderDLQProducer,
	handler *OrderEventHandler,
	batchSize int,
	batchTimeout time.Duration,
	workers int,
) *WorkerPool {
	return &WorkerPool{
		repo:         repo,
		dlqProducer:  dlqProducer,
		handler:      handler,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		workers:      workers,
	}
}
