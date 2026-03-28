package domain

import "time"

type Status string

const (
	StatusValid   Status = "valid"
	StatusDefect  Status = "defect"
	StatusWritten Status = "written_off"
)

type Container struct {
	ID          string
	TypeCode    string
	WarehouseID *string
	Status      Status
	CreatedAt   time.Time
}

type ContainerType struct {
	Code string
	Name string
}

type Warehouse struct {
	ID   string
	Name string
}

type ContainerFilter struct {
	TypeCode    *string
	WarehouseID *string
	Status      *Status

	Limit  int
	Offset int
}

func IsValidStatus(s Status) bool {
	switch s {
	case StatusValid, StatusDefect, StatusWritten:
		return true
	default:
		return false
	}
}
