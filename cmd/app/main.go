package main

import (
	"log"
	"net/http"

	"github.com/ljubushkin/container-management-service/internal/repository"
	httptransport "github.com/ljubushkin/container-management-service/internal/transport/http"
	"github.com/ljubushkin/container-management-service/internal/usecase"
)

func main() {
	repo := repository.NewInMemoryRepo()
	typeRepo := repository.NewTypeRepo()
	warehouseRepo := repository.NewWarehouseRepo()
	service := usecase.NewService(repo, typeRepo, warehouseRepo)
	handler := httptransport.NewHandler(service)

	router := httptransport.NewRouter(handler)

	log.Println("server started on :8080")
	http.ListenAndServe(":8080", router)
}
