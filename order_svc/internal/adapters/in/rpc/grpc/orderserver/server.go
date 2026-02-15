package orderserver

import (
	"net"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	pb "github.com/Anacardo89/order_svc_hex/order_svc/proto/orderpb"
	"google.golang.org/grpc"
)

// Server
type OrderGRPCServer struct {
	pb.UnimplementedOrderServiceServer
	Server   *grpc.Server
	Listener net.Listener
	service  ports.OrderGRPC
}

func NewOrderGRPCServer(port string, service ports.OrderGRPC) (*OrderGRPCServer, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	server := &OrderGRPCServer{
		Server:   s,
		Listener: listener,
		service:  service,
	}
	pb.RegisterOrderServiceServer(s, server)
	return server, nil
}

// Service
type OrderGRPCService struct {
	repo core.OrderRepo
}

func NewOrderGRPCService(repo core.OrderRepo) ports.OrderGRPC {
	return &OrderGRPCService{
		repo: repo,
	}
}
