package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ljubushkin/container-management-service/internal/repository"
	"github.com/ljubushkin/container-management-service/internal/usecase"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(s *usecase.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	var req CreateContainerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	c, err := h.service.CreateContainer(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := toResponse(c)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	c, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toResponse(c))
}

type AssignWarehouseRequest struct {
	WarehouseID string `json:"warehouse_id"`
}

func (h *Handler) AssignWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var req AssignWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := h.service.AssignWarehouse(id, req.WarehouseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
