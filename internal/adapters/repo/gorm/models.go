package gormrepo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Anacardo89/order_svc_hex/internal/core"
	"github.com/google/uuid"
)

type Order struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Items     Items     `gorm:"column:items;type:jsonb"`
	Status    string    `gorm:"column:status;type:order_status"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;autoUpdateTime"`
}

func fromCore(o *core.Order) *Order {
	return &Order{
		ID:        o.ID,
		Items:     o.Items,
		Status:    string(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func (o *Order) toCore() *core.Order {
	return &core.Order{
		ID:        o.ID,
		Items:     o.Items,
		Status:    core.Status(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

type Items map[string]int

// Implements driver.Valuer to write custom type
func (i Items) Value() (driver.Value, error) {
	if i == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(i)
}

// Implements sql.Scanner to read custom type
func (i *Items) Scan(value any) error {
	if value == nil {
		*i = make(Items)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Items: %T", value)
	}
	return json.Unmarshal(bytes, i)
}
