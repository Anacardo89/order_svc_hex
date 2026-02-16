package orderreader

import (
	"context"
	"io"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/proto/orderpb"
	"github.com/google/uuid"
)

func (c *OrderReaderClient) GetByID(ctx context.Context, id uuid.UUID) (*core.OrderResp, error) {
	resp, err := c.client.GetOrderByID(ctx, &orderpb.GetOrderByIDRequest{Id: id.String()})
	if err != nil {
		return nil, err
	}
	return fromProtoOrder(resp), nil
}

func (c *OrderReaderClient) ListByStatus(ctx context.Context, status core.Status) ([]*core.OrderResp, error) {
	stream, err := c.client.ListOrdersByStatus(ctx, &orderpb.ListOrdersByStatusRequest{Status: mapStatusToProto(status)})
	if err != nil {
		return nil, err
	}
	var orders []*core.OrderResp
	for {
		o, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		orders = append(orders, fromProtoOrder(o))
	}
	return orders, nil
}
