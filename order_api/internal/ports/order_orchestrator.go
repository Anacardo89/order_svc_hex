package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

type OrderOrchestrator interface {
	GetOrder(ctx context.Context, query *core.GetOrderQuery) (*core.Order, error)
	ListOrdersByStatus(ctx context.Context, query *core.ListOrdersByStatusQuery) ([]*core.Order, error)
	CreateOrder(ctx context.Context, req *core.CreateOrderCmd) error
	UpdateOrderStatus(ctx context.Context, req *core.UpdateOrderStatusCmd) error
}
