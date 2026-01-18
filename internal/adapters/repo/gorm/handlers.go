package gormrepo

import (
	"github.com/Anacardo89/order_svc_hex/internal/ports"
	"gorm.io/gorm"
)

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) ports.OrderRepo {
	return &OrderRepo{
		db: db,
	}
}
