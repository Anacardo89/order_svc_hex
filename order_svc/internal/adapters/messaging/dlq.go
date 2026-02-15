package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Anacardo89/order_svc_hex/order_svc/internal/core"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type dlqPayload struct {
	Timestamp     time.Time `json:"timestamp"`
	Reason        string    `json:"reason"`
	OriginalTopic string    `json:"original_topic"`
	OriginalKey   []byte    `json:"original_key,omitempty"`
	OriginalValue []byte    `json:"original_value"`
	Partition     int32     `json:"partition,omitempty"`
	Offset        int64     `json:"offset,omitempty"`
}

func (r *KafkaRawDLQProducer) PublishRawDLQ(ctx context.Context, msg *kafka.Message, reason string) error {
	payload := dlqPayload{
		Timestamp:     time.Now().UTC(),
		Reason:        reason,
		OriginalTopic: *msg.TopicPartition.Topic,
		OriginalKey:   msg.Key,
		OriginalValue: msg.Value,
		Partition:     msg.TopicPartition.Partition,
		Offset:        int64(msg.TopicPartition.Offset),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ payload: %w", err)
	}
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &r.topic,
			Partition: kafka.PartitionAny,
		},
		Value:     data,
		Key:       msg.Key,
		Timestamp: time.Now(),
	}
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)
	if err := r.producer.Produce(kafkaMsg, deliveryChan); err != nil {
		return fmt.Errorf("failed to produce DLQ message: %w", err)
	}
	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			slog.Error("[RawDLQProducer] failed delivery", "error", m.TopicPartition.Error)
			return m.TopicPartition.Error
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

type eventDlqPayload struct {
	Timestamp time.Time       `json:"timestamp"`
	Reason    string          `json:"reason"`
	Error     error           `json:"error"`
	Event     core.OrderEvent `json:"event"`
}

func (r *OrderDLQProducer) PublishDLQ(ctx context.Context, event core.OrderEvent, reason string, dlqErr error) error {
	payload := eventDlqPayload{
		Timestamp: time.Now().UTC(),
		Reason:    reason,
		Error:     dlqErr,
		Event:     event,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal OrderEvent DLQ payload: %w", err)
	}
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Partition: kafka.PartitionAny,
		},
		Value:     data,
		Key:       []byte(event.OrderID.String()),
		Timestamp: time.Now(),
	}
	switch event.EventType {
	case core.EventOrderCreated:
		kafkaMsg.TopicPartition.Topic = r.topics[string(core.EventOrderCreated)]
	case core.EventOrderStatusUpdated:
		kafkaMsg.TopicPartition.Topic = r.topics[string(core.EventOrderStatusUpdated)]
	}
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)
	if err := r.producer.Produce(kafkaMsg, deliveryChan); err != nil {
		return fmt.Errorf("failed to produce OrderEvent DLQ message: %w", err)
	}
	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			slog.Error("[OrderDLQProducer] failed delivery", "error", m.TopicPartition.Error)
			return m.TopicPartition.Error
		}
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
