package orderserver

import (
	"net"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	pb "github.com/Anacardo89/order_svc_hex/order_svc/proto/orderpb"
	"google.golang.org/grpc"
)

type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	Server   *grpc.Server
	Listener net.Listener
	service  ports.OrderServer
}

func NewOrderGRPCServer(port string, service ports.OrderServer) (*OrderGRPCServer, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(UnaryServerInterceptor()),
		grpc.StreamInterceptor(StreamServerInterceptor()),
	)
	server := &OrderGRPCServer{
		Server:   s,
		Listener: listener,
		service:  service,
	}
	pb.RegisterOrderServiceServer(s, server)
	return server, nil
}
