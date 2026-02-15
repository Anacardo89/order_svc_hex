package messaging

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
)

func (h *OrderEventHandler) OnOrderCreated(ctx context.Context, event core.OrderEvent) error {
	select {
	case h.createdQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *OrderEventHandler) OnOrderStatusUpdated(ctx context.Context, event core.OrderEvent) error {
	select {
	case h.statusUpdatedQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
