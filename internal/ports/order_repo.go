package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/internal/core"
	"github.com/google/uuid"
)

type OrderRepo interface {
	Create(ctx context.Context, order *core.Order) error
	GetByID(ctx context.Context, orderID uuid.UUID) (*core.Order, error)
	GetByStatus(ctx context.Context, status core.Status) ([]*core.Order, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status core.Status) error
}
