package grpctransport

import (
	"time"

	"github.com/ljubushkin/container-management-service/internal/domain"
	containerv1 "github.com/ljubushkin/container-management-service/pkg/api/container/v1"
)

func toProtoContainer(c *domain.Container) *containerv1.Container {
	if c == nil {
		return nil
	}

	var warehouseID string
	if c.WarehouseID != nil {
		warehouseID = *c.WarehouseID
	}

	return &containerv1.Container{
		Id:          c.ID,
		Type:        c.TypeCode,
		Status:      string(c.Status),
		WarehouseId: warehouseID,
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
	}
}

func toProtoContainers(containers []*domain.Container) []*containerv1.Container {
	resp := make([]*containerv1.Container, 0, len(containers))
	for _, c := range containers {
		resp = append(resp, toProtoContainer(c))
	}
	return resp
}
