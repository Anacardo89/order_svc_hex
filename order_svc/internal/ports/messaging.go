package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
)

// Consumer
type OrderConsumer interface {
	Consume(ctx context.Context, handler OrderEventHandler) error
}

type OrderEventHandler interface {
	OnOrderCreated(ctx context.Context, event core.OrderEvent) error
	OnOrderStatusUpdated(ctx context.Context, event core.OrderEvent) error
}

// Producer
type OrderDLQProducer interface {
	PublishDLQ(ctx context.Context, event core.OrderEvent, reason string, err error) error
}
