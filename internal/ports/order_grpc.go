package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/internal/core"
)

type OrderGRPC interface {
	GetOrderByID(ctx context.Context, id string) (*core.Order, error)
	GetOrdersByStatus(ctx context.Context, status core.Status) (<-chan *core.Order, error)
}
