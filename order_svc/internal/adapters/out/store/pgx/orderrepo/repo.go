package orderrepo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{
		pool: pool,
	}
}

func (r *OrderRepo) Close() {
	r.pool.Close()
}
