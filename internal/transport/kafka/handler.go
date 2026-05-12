package kafkatransport

import (
	"fmt"
	"log"

	"github.com/ljubushkin/container-management-service/internal/event"
	"github.com/ljubushkin/container-management-service/internal/usecase"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(service *usecase.Service) *Handler {
	return &Handler{service: service}
}

func (h Handler) HandleMovement(e event.ContainerMovement) error {
	switch e.EventType {
	case event.TypeContainerDispatched:
		log.Printf("handle dispatched event_id=%s container_id=%s trip_id=%s",
			e.EventID,
			e.ContainerID,
			e.TripID,
		)

		return h.service.UnassignWarehouse(e.ContainerID)
	case event.TypeContainerArrived:
		log.Printf("handle arrived event_id=%s container_id=%s warehouse_id=%s trip_id=%s",
			e.EventID,
			e.ContainerID,
			e.WarehouseID,
			e.TripID,
		)

		return h.service.AssignWarehouse(e.ContainerID, e.WarehouseID)
	default:
		return fmt.Errorf("unknown event type: %s", e.EventType)
	}
}
