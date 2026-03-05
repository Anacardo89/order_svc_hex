package orderserver

import (
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
)

type OrderGRPCService struct {
	repo ports.OrderRepo
}

func NewOrderGRPCService(repo ports.OrderRepo) ports.OrderServer {
	return &OrderGRPCService{
		repo: repo,
	}
}
