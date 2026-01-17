package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/internal/core"
)

type OrderConsumer interface {
	ConsumeOrderCreated(ctx context.Context, events chan<- core.OrderEvent) error
	ConsumeOrderStatusUpdated(ctx context.Context, events chan<- core.OrderEvent) error
}
