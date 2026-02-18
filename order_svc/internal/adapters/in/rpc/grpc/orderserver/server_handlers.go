package orderserver

import (
	"context"
	"fmt"

	pb "github.com/Anacardo89/order_svc_hex/order_svc/proto/orderpb"
)

func (s *OrderGRPCServer) GetOrderByID(ctx context.Context, req *pb.GetOrderByIDRequest) (*pb.Order, error) {
	order, err := s.service.GetOrderByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return toProtoOrder(order), nil
}

func (s *OrderGRPCServer) ListOrdersByStatus(req *pb.ListOrdersByStatusRequest, stream pb.OrderService_ListOrdersByStatusServer) error {
	orderCh, err := s.service.ListOrdersByStatus(stream.Context(), mapStatusToCore(req.Status))
	if err != nil {
		return err
	}
	for order := range orderCh {
		if err := stream.Send(toProtoOrder(order)); err != nil {
			return err
		}
	}
	return nil
}
