package orderorchestrator

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/internal/ports"
)

type OrderService struct {
	reader ports.OrderReader
	writer ports.OrderWriter
}

func NewOrderService(reader ports.OrderReader, writer ports.OrderWriter) *OrderService {
	return &OrderService{
		reader: reader,
		writer: writer,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, qry *core.GetOrderQry) (*core.Order, error) {
	return s.reader.GetByID(ctx, qry)
}

func (s *OrderService) ListOrdersByStatus(ctx context.Context, qry *core.ListOrdersByStatusQry) ([]*core.Order, error) {
	return s.reader.ListByStatus(ctx, qry)
}

func (s *OrderService) CreateOrder(ctx context.Context, cmd *core.CreateOrderCmd) error {
	return s.writer.PublishCreate(ctx, cmd)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, cmd *core.UpdateOrderStatusCmd) error {
	return s.writer.PublishStatusUpdate(ctx, cmd)
}
