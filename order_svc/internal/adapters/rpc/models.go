package rpc

import (
	pb "github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/rpc/orderpb"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/ptr"
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
		Status:    mapStatusToProto(ptr.Val(order.Status)),
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}
}
