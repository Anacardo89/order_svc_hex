package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

type OrderReader interface {
	GetByID(ctx context.Context, qry *core.GetOrderQry) (*core.Order, error)
	ListByStatus(ctx context.Context, qry *core.ListOrdersByStatusQry) ([]*core.Order, error)
}
