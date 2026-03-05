package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
)

type OrderRepo interface {
	Create(ctx context.Context, order *core.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*core.Order, error)
	ListByStatus(ctx context.Context, status core.Status) ([]*core.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status core.Status) error
}
