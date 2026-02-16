package orderserver

import (
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
)

type OrderGRPCService struct {
	repo core.OrderRepo
}

func NewOrderGRPCService(repo core.OrderRepo) ports.OrderServer {
	return &OrderGRPCService{
		repo: repo,
	}
}
