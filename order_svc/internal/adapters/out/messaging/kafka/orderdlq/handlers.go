package orderdlq

import (
	"context"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
)

func (c *DlqClient) PublishDLQ(ctx context.Context, msg ports.DLQMessage) error {
	var err string
	if msg.Error != nil {
		err = msg.Error.Error()
	}
	key := msg.OriginalKey
	if len(key) == 0 {
		key = []byte(msg.OriginalTopic)
	}
	payload := DlqPayload{
		Timestamp:     time.Now().UTC(),
		Reason:        msg.Reason,
		Error:         err,
		OriginalTopic: msg.OriginalTopic,
		OriginalKey:   msg.OriginalKey,
		OriginalValue: msg.OriginalValue,
		Partition:     msg.Partition,
		Offset:        msg.Offset,
	}
	return c.producerDlq.publish(ctx, string(key), payload)
}
