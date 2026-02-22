package orderrepo

import (
	"context"
	"encoding/json"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/adapters/infra/log/loki/logger"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("order_svc.postgres")
)

func (r *OrderRepo) Create(ctx context.Context, order *core.Order) error {
	// Observability
	ctx, span := tracer.Start(ctx, "db.orders.create",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "INSERT"),
			attribute.String("db.sql.table", "orders"),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
	query := `
		INSERT INTO orders (
			id,
			items,
			status
		)
		VALUES (
			$1, 
			$2,
			COALESCE($3::order_status, 'pending'::order_status)
		)
	;`
	dbOrder := fromCore(order)
	if dbOrder.ID == uuid.Nil {
		dbOrder.ID = uuid.New()
	}
	items, err := json.Marshal(dbOrder.Items)
	if err != nil {
		log.Error(ctx, "marshal items failed", ports.Field{Key: "error", Value: err})
		return failExec(span, "marshal items failed", err)
	}
	tag, err := r.pool.Exec(ctx, query, dbOrder.ID, items, dbOrder.Status)
	if err != nil {
		log.Error(ctx, "query failed", ports.Field{Key: "error", Value: err})
		return failExec(span, "query failed", err)
	}
	span.SetAttributes(attribute.Int64("db.rows_affected", tag.RowsAffected()))
	return nil
}

func (r *OrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*core.Order, error) {
	// Observability
	ctx, span := tracer.Start(ctx, "db.orders.get_by_id",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.sql.table", "orders"),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
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
		log.Error(ctx, "scan failed", ports.Field{Key: "error", Value: err})
		return failQueryRow[core.Order](span, "scan failed", err)
	}
	if err := json.Unmarshal(items, &dbOrder.Items); err != nil {
		log.Error(ctx, "unmarshal items failed", ports.Field{Key: "error", Value: err})
		return failQueryRow[core.Order](span, "unmarshal items failed", err)
	}
	dbOrder.Status = &status
	return dbOrder.toCore(), nil
}

func (r *OrderRepo) ListByStatus(ctx context.Context, status core.Status) ([]*core.Order, error) {
	// Observability
	ctx, span := tracer.Start(ctx, "db.orders.list_by_status",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.sql.table", "orders"),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
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
		log.Error(ctx, "query failed", ports.Field{Key: "error", Value: err})
		return failQuery[core.Order](span, "query failed", err)
	}
	defer rows.Close()
	var (
		count  int
		orders []*core.Order
	)
	for rows.Next() {
		count++
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
			log.Error(ctx, "scan failed", ports.Field{Key: "error", Value: err})
			return failQuery[core.Order](span, "scan failed", err)
		}
		if err := json.Unmarshal(items, &dbOrder.Items); err != nil {
			log.Error(ctx, "unmarshal items failed", ports.Field{Key: "error", Value: err})
			return failQuery[core.Order](span, "unmarshal items failed", err)
		}
		dbOrder.Status = &status
		orders = append(orders, dbOrder.toCore())
	}
	if err := rows.Err(); err != nil {
		log.Error(ctx, "rows loop failed", ports.Field{Key: "error", Value: err})
		return failQuery[core.Order](span, "rows loop failed", err)
	}
	span.SetAttributes(attribute.Int("db.rows_returned", count))
	return orders, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status core.Status) error {
	// Observability
	ctx, span := tracer.Start(ctx, "db.orders.update_status",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "UPDATE"),
			attribute.String("db.sql.table", "orders"),
		),
	)
	log := logger.LogFromSpan(span, logger.BaseLogger)
	defer span.End()

	// Execution
	query := `
		UPDATE orders
		SET status = $2
		WHERE id = $1
	;`
	tag, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		log.Error(ctx, "query failed", ports.Field{Key: "error", Value: err})
		return failExec(span, "query failed", err)
	}
	affected := tag.RowsAffected()
	span.SetAttributes(attribute.Int64("db.rows_affected", affected))
	if affected == 0 {
		span.SetStatus(codes.Error, "no rows updated")
	}
	return nil
}
