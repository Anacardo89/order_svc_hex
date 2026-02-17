package core

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrder struct {
	Items  map[string]int `json:"items" validate:"required"`
	Status Status         `json:"status"`
}

type UpdateOrderStatus struct {
	ID     string `json:"id" validate:"required"`
	Status Status `json:"status" validate:"required"`
}

type Order struct {
	ID        uuid.UUID      `json:"id"`
	Items     map[string]int `json:"items"`
	Status    Status         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
