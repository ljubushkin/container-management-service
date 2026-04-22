package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ljubushkin/container-management-service/internal/config"
	"github.com/ljubushkin/container-management-service/internal/repository"
	"github.com/ljubushkin/container-management-service/internal/repository/inmemory"
	"github.com/ljubushkin/container-management-service/internal/repository/postgres"
	grpctransport "github.com/ljubushkin/container-management-service/internal/transport/grpc"
	httptransport "github.com/ljubushkin/container-management-service/internal/transport/http"
	"github.com/ljubushkin/container-management-service/internal/usecase"
	containerv1 "github.com/ljubushkin/container-management-service/pkg/api/container/v1"
	"google.golang.org/grpc"
)

func mustOpenDB(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	return db
}

func buildRepositories(cfg config.Config) (
	repository.Repository,
	repository.ContainerTypeRepository,
	repository.WarehouseRepository,
	*sql.DB,
) {
	switch cfg.Storage {
	case "inmemory":
		log.Println("using inmemory repositories")

		return inmemory.NewContainerRepo(),
			inmemory.NewContainerTypeRepo(),
			inmemory.NewWarehouseRepo(),
			nil

	case "postgres":
		db := mustOpenDB(cfg.PostgresDSN)

		log.Println("using postgres repositories")

		return postgres.NewContainerRepo(db),
			postgres.NewContainerTypeRepo(db),
			postgres.NewWarehouseRepo(db),
			db

	default:
		log.Fatalf("unknown STORAGE value: %s", cfg.Storage)
		return nil, nil, nil, nil
	}
}

func main() {
	cfg := config.Load()

	containerRepo, typeRepo, warehouseRepo, db := buildRepositories(cfg)
	if db != nil {
		defer db.Close()
	}

	service := usecase.NewService(containerRepo, typeRepo, warehouseRepo)

	handler := httptransport.NewHandler(service)
	router := httptransport.NewRouter(handler)

	grpcHandler := grpctransport.NewServer(service)
	grpcServer := grpc.NewServer()
	containerv1.RegisterContainerServiceServer(grpcServer, grpcHandler)

	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("grpc listen: %v", err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server started on :%s", cfg.HTTPPort)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	go func() {
		log.Printf("grpc server started on :%s", cfg.GRPCPort)

		if err := grpcServer.Serve(grpcLis); err != nil {
			log.Printf("grpc server stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
		if err := server.Close(); err != nil {
			log.Printf("force close failed: %v", err)
		}
	}

	grpcServer.GracefulStop()

	log.Println("server stopped")
}
