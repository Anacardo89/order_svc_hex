package ports

import (
	"context"
	"encoding/json"
)

type OrderDLQ interface {
	PublishDLQ(ctx context.Context, payload json.RawMessage, reason string, err error) error
}
