package grpc

import (
	pb "github.com/Anacardo89/order_svc_hex/internal/adapters/api/orderpb"
	"github.com/Anacardo89/order_svc_hex/internal/core"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapStatusToProto(status core.Status) pb.OrderStatus {
	switch status {
	case core.StatusPending:
		return pb.OrderStatus_STATUS_PENDING
	case core.StatusConfirmed:
		return pb.OrderStatus_STATUS_CONFIRMED
	case core.StatusFailed:
		return pb.OrderStatus_STATUS_FAILED
	default:
		return pb.OrderStatus_STATUS_PENDING
	}
}

func mapProtoStatusToCore(s pb.OrderStatus) core.Status {
	switch s {
	case pb.OrderStatus_STATUS_PENDING:
		return core.StatusPending
	case pb.OrderStatus_STATUS_CONFIRMED:
		return core.StatusConfirmed
	case pb.OrderStatus_STATUS_FAILED:
		return core.StatusFailed
	default:
		return ""
	}
}

func toProtoOrder(order *core.Order) *pb.Order {
	items := make(map[string]int32, len(order.Items))
	for k, v := range order.Items {
		items[k] = int32(v)
	}
	return &pb.Order{
		Id:        order.ID.String(),
		Items:     items,
		Status:    mapStatusToProto(order.Status),
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}
}
