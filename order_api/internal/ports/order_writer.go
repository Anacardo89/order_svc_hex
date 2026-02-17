package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

type OrderWriter interface {
	PublishCreate(ctx context.Context, req *core.CreateOrder) error
	PublishStatusUpdate(ctx context.Context, req *core.UpdateOrderStatus) error
}
