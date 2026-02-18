package orderreader

import (
	pb "github.com/Anacardo89/order_svc_hex/order_api/proto/orderpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderReaderClient struct {
	client pb.OrderServiceClient
	conn   *grpc.ClientConn
}

func NewOrderReaderClient(addr string) (*OrderReaderClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := pb.NewOrderServiceClient(conn)
	return &OrderReaderClient{
		client: c,
		conn:   conn,
	}, nil
}

func (c *OrderReaderClient) Close() error {
	return c.conn.Close()
}
