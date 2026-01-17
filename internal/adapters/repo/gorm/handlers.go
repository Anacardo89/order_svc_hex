package gormrepo

import "gorm.io/gorm"

type OrderRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}
