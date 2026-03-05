package core

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID      `json:"id"`
	Items     map[string]int `json:"items"`
	Status    Status         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Commands
type CreateOrderCmd struct {
	Items  map[string]int `json:"items" validate:"required"`
	Status Status         `json:"status"`
}

type UpdateOrderStatusCmd struct {
	ID     string `json:"id" validate:"required"`
	Status Status `json:"status" validate:"required"`
}

// Queries
type GetOrderQuery struct {
	ID uuid.UUID
}

type ListOrdersByStatusQuery struct {
	Status Status
}
