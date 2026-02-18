package orderreader

import (
	"fmt"

	"github.com/Anacardo89/order_svc_hex/order_api/config"
	pb "github.com/Anacardo89/order_svc_hex/order_api/proto/orderpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderReaderClient struct {
	client pb.OrderServiceClient
	conn   *grpc.ClientConn
}

func NewOrderReaderClient(cfg config.GRPC) (*OrderReaderClient, error) {
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
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
