package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ljubushkin/container-management-service/internal/repository/postgres"
	httptransport "github.com/ljubushkin/container-management-service/internal/transport/http"
	"github.com/ljubushkin/container-management-service/internal/usecase"
)

func mustOpenDB(dsn string) *sql.DB {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}

func main() {
	db := mustOpenDB("postgres://postgres:postgres@localhost:5432/containers?sslmode=disable")

	containerRepo := postgres.NewContainerRepo(db)
	typeRepo := postgres.NewContainerTypeRepo(db)
	warehouseRepo := postgres.NewWarehouseRepo(db)

	service := usecase.NewService(containerRepo, typeRepo, warehouseRepo)

	// repo := inmemory.NewContainerRepo()
	// typeRepo := inmemory.NewContainerTypeRepo()
	// warehouseRepo := inmemory.NewWarehouseRepo()
	// service := usecase.NewService(repo, typeRepo, warehouseRepo)
	handler := httptransport.NewHandler(service)

	router := httptransport.NewRouter(handler)

	log.Println("server started on :8080")
	http.ListenAndServe(":8080", router)
}
