package orderorchestrator

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
	"github.com/google/uuid"
)

type OrderService struct {
	reader ports.OrderReader
	writer ports.OrderWriter
}

func NewOrderService(reader ports.OrderReader, writer ports.OrderWriter) core.OrderOrchestrator {
	return &OrderService{
		reader: reader,
		writer: writer,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*core.Order, error) {
	return s.reader.GetByID(ctx, id)
}

func (s *OrderService) ListOrdersByStatus(ctx context.Context, status *core.Status) ([]*core.Order, error) {
	return s.reader.ListByStatus(ctx, *status)
}

func (s *OrderService) CreateOrder(ctx context.Context, req *core.CreateOrder) error {
	return s.writer.PublishCreate(ctx, req)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, req *core.UpdateOrderStatus) error {
	return s.writer.PublishStatusUpdate(ctx, req)
}
