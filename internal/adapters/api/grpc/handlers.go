package grpc

import (
	"net"

	pb "github.com/Anacardo89/order_svc_hex/internal/adapters/api/orderpb"
	"github.com/Anacardo89/order_svc_hex/internal/ports"
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
	repo ports.OrderRepo
}

func NewOrderGRPCService(repo ports.OrderRepo) ports.OrderGRPC {
	return &OrderGRPCService{
		repo: repo,
	}
}
