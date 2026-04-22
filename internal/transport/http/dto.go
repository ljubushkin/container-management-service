package httptransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ljubushkin/container-management-service/internal/apperror"
	"github.com/ljubushkin/container-management-service/internal/domain"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type CreateContainerRequest struct {
	Type string `json:"type" validate:"required"`
}

type ContainerResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	WarehouseID *string `json:"warehouse_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type AssignWarehouseRequest struct {
	WarehouseID string `json:"warehouse_id" validate:"required"`
}

func toResponse(c *domain.Container) ContainerResponse {
	return ContainerResponse{
		ID:          c.ID,
		Type:        c.TypeCode,
		Status:      string(c.Status),
		WarehouseID: c.WarehouseID,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}
}

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError

	if errors.As(err, &appErr) {
		status := mapStatus(appErr.Code)

		resp := ErrorResponse{}
		resp.Error.Code = string(appErr.Code)
		resp.Error.Message = appErr.Message

		writeJSON(w, status, resp)
		return
	}

	// fallback
	resp := ErrorResponse{}
	resp.Error.Code = "INTERNAL_ERROR"
	resp.Error.Message = "internal error"

	writeJSON(w, http.StatusInternalServerError, resp)
}

func mapStatus(code apperror.Code) int {
	switch code {
	case apperror.CodeInvalidType,
		apperror.CodeInvalidStatus,
		apperror.CodeInvalidWarehouse:
		return http.StatusBadRequest

	case apperror.CodeNotFound:
		return http.StatusNotFound

	default:
		return http.StatusInternalServerError
	}
}

type CreateBatchRequest struct {
	Type  string `json:"type" validate:"required"`
	Count int    `json:"count" validate:"required,gt=0,lte=1000"`
}

type ListContainersResponse struct {
	Data []ContainerResponse `json:"data"`
	Meta Meta                `json:"meta"`
}

type Meta struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
}

type WarehouseResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ContainerTypeResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func toWarehouseResponse(w *domain.Warehouse) WarehouseResponse {
	return WarehouseResponse{
		ID:   w.ID,
		Name: w.Name,
	}
}

func toContainerTypeResponse(t *domain.ContainerType) ContainerTypeResponse {
	return ContainerTypeResponse{
		Code: t.Code,
		Name: t.Name,
	}
}
