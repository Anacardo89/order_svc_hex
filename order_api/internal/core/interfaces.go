package core

import (
	"context"

	"github.com/google/uuid"
)

type OrderOrchestrator interface {
	CreateOrder(ctx context.Context, req *CreateOrderReq) error
	GetOrder(ctx context.Context, id uuid.UUID) (*OrderResp, error)
	ListOrdersByStatus(ctx context.Context, status string) ([]*OrderResp, error)
	UpdateOrderStatus(ctx context.Context, req *UpdateOrderStatusReq) error
}
