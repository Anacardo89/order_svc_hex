package events

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConnection struct {
	Brokers string
}

func NewKafkaConnection(brokers string) *KafkaConnection {
	return &KafkaConnection{
		Brokers: brokers,
	}
}

func EnsureTopics(brokers string, topics []string, partition int) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return fmt.Errorf("failed to create Kafka admin client: %w", err)
	}
	defer admin.Close()

	var specs []kafka.TopicSpecification
	for _, topic := range topics {
		specs = append(specs, kafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     partition,
			ReplicationFactor: 1,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := admin.CreateTopics(ctx, specs)
	if err != nil {
		return fmt.Errorf("failed to create topics: %w", err)
	}

	for _, res := range results {
		if res.Error.Code() != kafka.ErrNoError && res.Error.Code() != kafka.ErrTopicAlreadyExists {
			return fmt.Errorf("failed to create topic %s: %v", res.Topic, res.Error)
		}
		fmt.Printf("Topic %s creation result: %v\n", res.Topic, res.Error)
	}

	return nil
}

func (c *KafkaConnection) MakeConsumer(groupID string, topics []string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	err = consumer.SubscribeTopics(topics, rebalanceCb)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topics: %w", err)
	}
	return consumer, nil
}

func (c *KafkaConnection) MakeProducer() (*kafka.Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": c.Brokers,
	})
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func rebalanceCb(c *kafka.Consumer, e kafka.Event) error {
	switch ev := e.(type) {
	case kafka.AssignedPartitions:
		fmt.Println("Partitions assigned:", ev.Partitions)
		c.Assign(ev.Partitions)
	case kafka.RevokedPartitions:
		fmt.Println("Partitions revoked:", ev.Partitions)
		c.Unassign()
	}
	return nil
}
