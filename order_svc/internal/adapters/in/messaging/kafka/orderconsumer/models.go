package orderconsumer

import (
	"encoding/json"
	"fmt"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/Anacardo89/order_svc_hex/order_svc/internal/ports"
	"github.com/Anacardo89/order_svc_hex/order_svc/pkg/ptr"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

type TopicKey string

const (
	TopicOrderCreated       TopicKey = "OrderCreated"
	TopicOrderStatusUpdated TopicKey = "OrderStatusUpdated"
)

type OrderCreatedEvent struct {
	Items  map[string]int `json:"items"`
	Status string         `json:"status"`
}

type OrderStatusUpdatedEvent struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func mapEventPaylodToOrder(msg *kafka.Message) (*core.Order, error) {
	switch *msg.TopicPartition.Topic {
	case "orders.created":
		var e OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderCreated: %w", err)
		}
		var status *core.Status
		if e.Status != "" {
			s := core.Status(e.Status)
			status = &s
		}
		return &core.Order{
			Items:  e.Items,
			Status: status,
		}, nil
	case "orders.status_updated":
		var e OrderStatusUpdatedEvent
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OrderStatusUpdated: %w", err)
		}
		id, err := uuid.Parse(e.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in OrderStatusUpdated: %w", err)
		}
		return &core.Order{
			ID:     id,
			Status: ptr.Ptr(core.Status(e.Status)),
		}, nil
	default:
		return nil, fmt.Errorf("unknown topic: %s", *msg.TopicPartition.Topic)
	}
}

func makeDlqMessage(msg *kafka.Message, reason string, err error) ports.DLQMessage {
	return ports.DLQMessage{
		Reason:        reason,
		Error:         err,
		OriginalTopic: *msg.TopicPartition.Topic,
		OriginalKey:   msg.Key,
		OriginalValue: msg.Value,
		Partition:     msg.TopicPartition.Partition,
		Offset:        int64(msg.TopicPartition.Offset),
	}
}
