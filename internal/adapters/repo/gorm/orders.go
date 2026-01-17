package gormrepo

import (
	"context"

	"github.com/Anacardo89/order_svc_hex/internal/core"
	"github.com/google/uuid"
)

func (r *OrderRepo) Create(ctx context.Context, order *core.Order) error {
	dbOrder := fromCore(order)
	return r.db.WithContext(ctx).
		Create(dbOrder).Error
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*core.Order, error) {
	var dbOrder Order
	if err := r.db.WithContext(ctx).
		First(&dbOrder, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return dbOrder.toCore(), nil
}

func (r *OrderRepo) GetByStatus(ctx context.Context, status core.Status) ([]*core.Order, error) {
	var dbOrders []Order
	if err := r.db.WithContext(ctx).
		Where("status = ?", string(status)).
		Find(&dbOrders).Error; err != nil {
		return nil, err
	}
	orders := make([]*core.Order, len(dbOrders))
	for i, o := range dbOrders {
		orders[i] = o.toCore()
	}
	return orders, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status core.Status) error {
	return r.db.WithContext(ctx).
		Model(&Order{}).
		Where("id = ?", id).
		Update("status", string(status)).Error
}
