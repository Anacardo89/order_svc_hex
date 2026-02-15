package orderreader

import (
	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	pb "github.com/Anacardo89/order_svc_hex/order_api/proto/orderpb"
	"github.com/google/uuid"
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

func mapStatusToCore(s pb.OrderStatus) core.Status {
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

func fromProtoOrder(o *pb.Order) *core.OrderResp {
	items := make(map[string]int, len(o.Items))
	for k, v := range o.Items {
		items[k] = int(v)
	}
	return &core.OrderResp{
		ID:        uuid.MustParse(o.Id),
		Items:     items,
		Status:    mapStatusToCore(o.Status),
		CreatedAt: o.CreatedAt.AsTime(),
		UpdatedAt: o.UpdatedAt.AsTime(),
	}
}
