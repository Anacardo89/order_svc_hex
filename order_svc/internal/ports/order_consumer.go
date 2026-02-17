package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
)

type OrderConsumer interface {
	OnOrderCreated(ctx context.Context, order core.Order) error
	OnOrderStatusUpdated(ctx context.Context, order core.Order) error
}
