package orderrepo

import (
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) core.OrderRepo {
	return &OrderRepo{
		pool: pool,
	}
}
