package orderconsumer

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
)

type OrderHandler struct {
	repo core.OrderRepo
}

func NewOrderHandler(repo core.OrderRepo) ports.OrderConsumer {
	return &OrderHandler{
		repo: repo,
	}
}

func (h *OrderHandler) OnOrderCreated(ctx context.Context, order core.Order) error {
	return h.repo.Create(ctx, &order)
}

func (h *OrderHandler) OnOrderStatusUpdated(ctx context.Context, order core.Order) error {
	return h.repo.UpdateStatus(ctx, order.ID, *order.Status)
}
