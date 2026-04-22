package config

import "os"

const (
	defaultHTTPPort    = "8080"
	defaultGRPCPort    = "50051"
	defaultStorage     = "postgres"
	defaultPostgresDSN = "postgres://postgres:postgres@localhost:5432/containers?sslmode=disable"
)

type Config struct {
	HTTPPort    string
	GRPCPort    string
	Storage     string
	PostgresDSN string
}

func Load() Config {
	return Config{
		HTTPPort:    getEnv("HTTP_PORT", defaultHTTPPort),
		GRPCPort:    getEnv("GRPC_PORT", defaultGRPCPort),
		Storage:     getEnv("STORAGE", defaultStorage),
		PostgresDSN: getEnv("POSTGRES_DSN", defaultPostgresDSN),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
