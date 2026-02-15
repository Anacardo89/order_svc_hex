package core

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusFailed    Status = "failed"
)

type Order struct {
	ID        uuid.UUID
	Items     map[string]int
	Status    *Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EventType string

const (
	EventOrderCreated       EventType = "OrderCreated"
	EventOrderStatusUpdated EventType = "OrderStatusUpdated"
)

type OrderEvent struct {
	OrderID   uuid.UUID      `json:"order_id"`
	Items     map[string]int `json:"items"`
	Status    *Status        `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	EventType EventType      `json:"event_type"`
}

func (e *OrderEvent) ToOrder() *Order {
	return &Order{
		ID:        e.OrderID,
		Items:     e.Items,
		Status:    e.Status,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}
