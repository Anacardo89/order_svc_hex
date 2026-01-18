package grpc

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/internal/core"
	"github.com/google/uuid"
)

func (s *OrderGRPCService) GetOrderByID(ctx context.Context, id string) (*core.Order, error) {
	orderID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderGRPCService) GetOrdersByStatus(ctx context.Context, status core.Status) (<-chan *core.Order, error) {
	orders, err := s.repo.GetByStatus(ctx, status)
	if err != nil {
		return nil, err
	}
	out := make(chan *core.Order, len(orders))
	go func() {
		defer close(out)
		for _, o := range orders {
			select {
			case <-ctx.Done():
				return
			case out <- o:
			}
		}
	}()
	return out, nil
}
