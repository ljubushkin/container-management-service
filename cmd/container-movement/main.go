package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	containerv1 "github.com/ljubushkin/container-management-service/pkg/api/container/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	EventContainerDispatched = "container.dispatched"
	EventContainerArrived    = "container.arrived"
	TopicContainerMovements  = "container.movements"
)

type ContainerMovement struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	OccurredAt  time.Time `json:"occurred_at"`
	ContainerID string    `json:"container_id"`
	WarehouseID string    `json:"warehouse_id"`
	TripID      string    `json:"trip_id"`
}

func getContainers(client containerv1.ContainerServiceClient, ctx context.Context) ([]*containerv1.Container, error) {
	resp, err := client.ListContainers(ctx, &containerv1.ListContainersRequest{})
	if err != nil {
		return nil, err
	}

	return resp.GetContainers(), nil
}

func getWarehouses(client containerv1.ContainerServiceClient, ctx context.Context) ([]*containerv1.Warehouse, error) {
	resp, err := client.ListWarehouses(ctx, &containerv1.ListWarehousesRequest{})
	if err != nil {
		return nil, err
	}

	return resp.GetWarehouses(), nil
}

func newID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func movement(client containerv1.ContainerServiceClient, producer *Producer) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := getContainers(client, ctx)
	if err != nil {
		log.Println(err)
		return
	}

	if len(containers) == 0 {
		log.Println("No available containers")
		return
	}

	warehouses, err := getWarehouses(client, ctx)
	if err != nil {
		log.Println(err)
		return
	}

	warehousesCount := len(warehouses)
	if warehousesCount == 0 {
		log.Println("No available warehouses")
		return
	}
	for _, c := range containers {
		ID := c.GetId()

		var event ContainerMovement

		switch {
		case c.GetWarehouseId() == "":
			w := warehouses[rand.Intn(warehousesCount)]
			event = ContainerMovement{
				EventID:     newID(),
				EventType:   EventContainerArrived,
				OccurredAt:  time.Now(),
				ContainerID: ID,
				WarehouseID: w.GetWarehouseId(),
				TripID:      "",
			}
		default:
			event = ContainerMovement{
				EventID:     newID(),
				EventType:   EventContainerDispatched,
				OccurredAt:  time.Now(),
				ContainerID: ID,
				WarehouseID: c.GetWarehouseId(),
				TripID:      newID(),
			}
		}
		message, err := json.Marshal(event)
		if err != nil {
			log.Printf("failed marshal message: %v", err)
			continue
		}
		if err := producer.Produce(message, TopicContainerMovements, ID); err != nil {
			log.Printf("failed produce event: %v", err)
			continue
		}

		time.Sleep(time.Millisecond * 500)
	}
}

func main() {

	conn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := containerv1.NewContainerServiceClient(conn)

	producer, err := NewProducer("localhost:9091,localhost:9092,localhost:9093")
	if err != nil {
		log.Fatalf("create producer: %v", err)
	}
	defer producer.Close()

	for {
		movement(client, producer)
		time.Sleep(time.Minute * 5)
	}

}

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(address string) (*Producer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": address,
	}
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed create producer: %w", err)
	}
	return &Producer{
		producer: producer,
	}, nil
}

func (p *Producer) Produce(message []byte, topic, key string) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: message,
		Key:   []byte(key),
	}
	ch := make(chan kafka.Event)
	defer close(ch)
	if err := p.producer.Produce(kafkaMsg, ch); err != nil {
		return err
	}
	e := <-ch
	msg, ok := e.(*kafka.Message)
	if !ok {
		return fmt.Errorf("unexpected kafka event: %T", e)
	}

	if msg.TopicPartition.Error != nil {
		return msg.TopicPartition.Error
	}
	log.Printf(
		"produced message topic=%s partition=%d offset=%v key=%s",
		*msg.TopicPartition.Topic,
		msg.TopicPartition.Partition,
		msg.TopicPartition.Offset,
		key,
	)
	return nil
}

func (p *Producer) Close() {
	remaining := p.producer.Flush(5000)
	if remaining > 0 {
		log.Printf("producer closed with %d undelivered messages", remaining)
	}

	p.producer.Close()
}
