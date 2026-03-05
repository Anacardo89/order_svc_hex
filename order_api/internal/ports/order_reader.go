package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

type OrderReader interface {
	GetByID(ctx context.Context, query *core.GetOrderQuery) (*core.Order, error)
	ListByStatus(ctx context.Context, query *core.ListOrdersByStatusQuery) ([]*core.Order, error)
}
