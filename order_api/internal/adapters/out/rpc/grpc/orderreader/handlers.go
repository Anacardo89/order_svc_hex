package orderreader

import (
	"context"
	"io"

	"github.com/Anacardo89/order_svc_hex/order_api/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_api/proto/orderpb"
)

func (c *OrderReaderClient) GetByID(ctx context.Context, qry *core.GetOrderQry) (*core.Order, error) {
	resp, err := c.client.GetOrderByID(ctx, &orderpb.GetOrderByIDRequest{Id: qry.ID.String()})
	if err != nil {
		return nil, err
	}
	return fromProtoOrder(resp), nil
}

func (c *OrderReaderClient) ListByStatus(ctx context.Context, qry *core.ListOrdersByStatusQry) ([]*core.Order, error) {
	stream, err := c.client.ListOrdersByStatus(ctx, &orderpb.ListOrdersByStatusRequest{Status: mapStatusToProto(qry.Status)})
	if err != nil {
		return nil, err
	}
	var orders []*core.Order
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
