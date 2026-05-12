package kafkatransport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/ljubushkin/container-management-service/internal/event"
)

const sessionTimeoutMs = 7000

type Consumer struct {
	consumer *kafka.Consumer
	topic    string
	handler  *Handler
}

func NewConsumer(address, topic, groupID string, handler *Handler) (*Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        address,
		"group.id":                 groupID,
		"session.timeout.ms":       sessionTimeoutMs,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"auto.offset.reset":        "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("subscribe topic %s: %w", topic, err)
	}

	return &Consumer{
		consumer: c,
		topic:    topic,
		handler:  handler,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	log.Printf("kafka consumer started topic=%s", c.topic)

	defer func() {
		if err := c.Stop(); err != nil {
			log.Printf("kafka consumer stop: %v", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("kafka consumer shutdown requested")
			return
		default:
		}

		msg, err := c.consumer.ReadMessage(1 * time.Second)
		if err != nil {
			// timeout чтения — это нормально
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			}

			log.Printf("read kafka message: %v", err)
			continue
		}

		if msg == nil {
			continue
		}

		if err := c.handleMessage(msg); err != nil {
			log.Printf("handle kafka message: %v", err)
			continue
		}

		if _, err := c.consumer.StoreMessage(msg); err != nil {
			log.Printf("store kafka offset: %v", err)
			continue
		}
	}
}

func (c *Consumer) handleMessage(msg *kafka.Message) error {
	var movement event.ContainerMovement

	if err := json.Unmarshal(msg.Value, &movement); err != nil {
		return fmt.Errorf("unmarshal movement event: %w", err)
	}

	if movement.ContainerID == "" {
		return fmt.Errorf("container_id is required")
	}

	if movement.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if err := c.handler.HandleMovement(movement); err != nil {
		return fmt.Errorf("handle movement event_id=%s type=%s: %w",
			movement.EventID,
			movement.EventType,
			err,
		)
	}

	log.Printf(
		"processed kafka message topic=%s partition=%d offset=%v key=%s",
		*msg.TopicPartition.Topic,
		msg.TopicPartition.Partition,
		msg.TopicPartition.Offset,
		string(msg.Key),
	)

	return nil
}

func (c *Consumer) Stop() error {
	if _, err := c.consumer.Commit(); err != nil {
		return fmt.Errorf("commit kafka offsets: %w", err)
	}

	log.Println("kafka offsets committed")

	if err := c.consumer.Close(); err != nil {
		return fmt.Errorf("close kafka consumer: %w", err)
	}

	return nil
}
