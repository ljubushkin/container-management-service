package repository

import "github.com/ljubushkin/container-management-service/internal/domain"

type Repository interface {
	Create(c *domain.Container) error
	CreateBatch(containers []*domain.Container) error
	GetByID(id string) (*domain.Container, error)
	List(filter domain.ContainerFilter) ([]*domain.Container, error)
	Update(c *domain.Container) error
}

type ContainerTypeRepository interface {
	GetByCode(code string) (*domain.ContainerType, error)
	List() ([]*domain.ContainerType, error)
}

type WarehouseRepository interface {
	GetByID(id string) (*domain.Warehouse, error)
	List() ([]*domain.Warehouse, error)
}
