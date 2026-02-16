package ports

import (
	"context"
)

type DLQMessage struct {
	Reason        string
	Error         error
	OriginalTopic string
	OriginalKey   []byte
	OriginalValue []byte
	Partition     int32
	Offset        int64
}

type OrderDLQ interface {
	PublishDLQ(ctx context.Context, msg DLQMessage) error
}
