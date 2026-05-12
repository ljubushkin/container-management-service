package event

import "time"

const (
	TopicContainerMovements = "container.movements"

	TypeContainerDispatched = "container.dispatched"
	TypeContainerArrived    = "container.arrived"
)

type ContainerMovement struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	OccurredAt  time.Time `json:"occurred_at"`
	ContainerID string    `json:"container_id"`
	WarehouseID string    `json:"warehouse_id"`
	TripID      string    `json:"trip_id"`
}
