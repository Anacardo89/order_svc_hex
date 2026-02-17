package orderrepo

import (
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID
	Items     map[string]int
	Status    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func fromCore(o *core.Order) *Order {
	var status *string
	if o.Status != nil {
		s := string(*o.Status)
		status = &s
	}
	return &Order{
		ID:        o.ID,
		Items:     o.Items,
		Status:    status,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func (o *Order) toCore() *core.Order {
	var status *core.Status
	if o.Status != nil {
		s := core.Status(*o.Status)
		status = &s
	}
	return &core.Order{
		ID:        o.ID,
		Items:     o.Items,
		Status:    status,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}
