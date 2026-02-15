package core

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepo interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	ListByStatus(ctx context.Context, status Status) ([]*Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
}
