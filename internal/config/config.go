package config

import "os"

const (
	defaultHTTPPort            = "8080"
	defaultGRPCPort            = "50051"
	defaultStorage             = "postgres"
	defaultPostgresDSN         = "postgres://postgres:postgres@localhost:5432/containers?sslmode=disable"
	defaultKafkaBrokers        = "localhost:9091,localhost:9092,localhost:9093"
	defaultKafkaMovementsTopic = "container.movements"
	defaultKafkaGroupID        = "container-management-movements"
)

type Config struct {
	HTTPPort            string
	GRPCPort            string
	Storage             string
	PostgresDSN         string
	KafkaBrokers        string
	KafkaMovementsTopic string
	KafkaGroupID        string
}

func Load() Config {
	return Config{
		HTTPPort:            getEnv("HTTP_PORT", defaultHTTPPort),
		GRPCPort:            getEnv("GRPC_PORT", defaultGRPCPort),
		Storage:             getEnv("STORAGE", defaultStorage),
		PostgresDSN:         getEnv("POSTGRES_DSN", defaultPostgresDSN),
		KafkaBrokers:        getEnv("KAFKA_BROKERS", defaultKafkaBrokers),
		KafkaMovementsTopic: getEnv("KAFKA_MOVEMENTS_TOPIC", defaultKafkaMovementsTopic),
		KafkaGroupID:        getEnv("KAFKA_GROUP_ID", defaultKafkaGroupID),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
