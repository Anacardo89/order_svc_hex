package repo

import (
	"context"
	"encoding/json"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/google/uuid"
)

func (r *OrderRepo) Create(ctx context.Context, order *core.Order) error {
	query := `
		INSERT INTO orders (
			id,
			items,
			status
		)
		VALUES ($1, $2, COALESCE($3::order_status, 'pending'::order_status))
	;`
	dbOrder := fromCore(order)
	if dbOrder.ID == uuid.Nil {
		dbOrder.ID = uuid.New()
	}
	items, err := json.Marshal(dbOrder.Items)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, query, dbOrder.ID, items, dbOrder.Status)
	return err
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*core.Order, error) {
	query := `
		SELECT
			id,
			items,
			status,
			created_at,
			updated_at
		FROM orders
		WHERE id = $1
	;`
	var (
		dbOrder Order
		items   []byte
		status  string
	)
	if err := r.pool.QueryRow(ctx, query, id).Scan(
		&dbOrder.ID,
		&items,
		&status,
		&dbOrder.CreatedAt,
		&dbOrder.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(items, &dbOrder.Items); err != nil {
		return nil, err
	}
	dbOrder.Status = &status
	return dbOrder.toCore(), nil
}

func (r *OrderRepo) ListByStatus(ctx context.Context, status core.Status) ([]*core.Order, error) {
	query := `
		SELECT
			id,
			items,
			status,
			created_at,
			updated_at
		FROM orders
		WHERE status = $1
	;`
	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []*core.Order
	for rows.Next() {
		var dbOrder Order
		var items []byte
		var status string
		if err := rows.Scan(
			&dbOrder.ID,
			&items,
			&status,
			&dbOrder.CreatedAt,
			&dbOrder.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(items, &dbOrder.Items); err != nil {
			return nil, err
		}
		dbOrder.Status = &status
		orders = append(orders, dbOrder.toCore())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status core.Status) error {
	query := `
		UPDATE orders
		SET status = $2
		WHERE id = $1
	;`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}
