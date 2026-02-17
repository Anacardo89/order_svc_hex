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
