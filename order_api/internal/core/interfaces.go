package core

import (
	"context"

	"github.com/google/uuid"
)

type OrderOrchestrator interface {
	GetOrder(ctx context.Context, id uuid.UUID) (*Order, error)
	ListOrdersByStatus(ctx context.Context, status *Status) ([]*Order, error)
	CreateOrder(ctx context.Context, req *CreateOrder) error
	UpdateOrderStatus(ctx context.Context, req *UpdateOrderStatus) error
}
