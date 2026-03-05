package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
)

type OrderWriter interface {
	PublishCreate(ctx context.Context, cmd *core.CreateOrderCmd) error
	PublishStatusUpdate(ctx context.Context, cmd *core.UpdateOrderStatusCmd) error
}
