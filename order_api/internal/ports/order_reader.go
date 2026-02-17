package ports

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/google/uuid"
)

type OrderReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*core.Order, error)
	ListByStatus(ctx context.Context, status core.Status) ([]*core.Order, error)
}
