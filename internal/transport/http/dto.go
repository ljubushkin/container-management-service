package httptransport

import (
	"time"

	"github.com/ljubushkin/container-management-service/internal/domain"
)

type CreateContainerRequest struct {
	Type string `json:"type"`
}

type ContainerResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	WarehouseID *string `json:"warehouse_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type AssignWarehouseRequest struct {
	WarehouseID string `json:"warehouse_id"`
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
