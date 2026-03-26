package repository

import (
	"errors"
	"sync"

	"github.com/ljubushkin/container-management-service/internal/domain"
)

type Repository interface {
	Create(c *domain.Container) error
	CreateBatch(containers []*domain.Container) error
	GetByID(id string) (*domain.Container, error)
	List(filter ListFilter) ([]*domain.Container, error)
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

type ListFilter struct {
	Type        *string
	WarehouseID *string
	Status      *domain.Status
	Limit       *int
	Offset      *int
}

var ErrNotFound = errors.New("container not found")

type InMemoryRepo struct {
	mu   sync.RWMutex
	data map[string]*domain.Container
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		data: make(map[string]*domain.Container),
	}
}

type InMemoryTypeRepo struct {
	data map[string]*domain.ContainerType
}

func NewTypeRepo() *InMemoryTypeRepo {
	return &InMemoryTypeRepo{
		data: map[string]*domain.ContainerType{
			"EURO_PALLET": {Code: "EURO_PALLET", Name: "Euro pallet"},
			"BOX":         {Code: "BOX", Name: "Box"},
		},
	}
}

func (r *InMemoryTypeRepo) GetByCode(code string) (*domain.ContainerType, error) {
	t, ok := r.data[code]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *InMemoryTypeRepo) List() ([]*domain.ContainerType, error) {
	result := make([]*domain.ContainerType, 0, len(r.data))
	for _, t := range r.data {
		result = append(result, t)
	}
	return result, nil
}

type InMemoryWarehouseRepo struct {
	data map[string]*domain.Warehouse
}

func NewWarehouseRepo() *InMemoryWarehouseRepo {
	return &InMemoryWarehouseRepo{
		data: map[string]*domain.Warehouse{
			"w1": {ID: "w1", Name: "Main"},
			"w2": {ID: "w2", Name: "Reserve"},
		},
	}
}

func (r *InMemoryWarehouseRepo) GetByID(id string) (*domain.Warehouse, error) {
	w, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return w, nil
}

func (r *InMemoryWarehouseRepo) List() ([]*domain.Warehouse, error) {
	result := make([]*domain.Warehouse, 0, len(r.data))
	for _, w := range r.data {
		result = append(result, w)
	}
	return result, nil
}

func copyContainer(c *domain.Container) *domain.Container {
	if c == nil {
		return nil
	}

	copy := *c

	if c.WarehouseID != nil {
		wid := *c.WarehouseID
		copy.WarehouseID = &wid
	}

	return &copy
}

func (r *InMemoryRepo) Create(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[c.ID] = copyContainer(c)
	return nil
}

func (r *InMemoryRepo) CreateBatch(containers []*domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range containers {
		r.data[c.ID] = copyContainer(c)
	}
	return nil
}

func (r *InMemoryRepo) GetByID(id string) (*domain.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}

	return copyContainer(c), nil
}

func (r *InMemoryRepo) List(filter ListFilter) ([]*domain.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Container, 0)

	for _, c := range r.data {
		if filter.Type != nil && c.TypeCode != *filter.Type {
			continue
		}

		if filter.Status != nil && c.Status != *filter.Status {
			continue
		}

		if filter.WarehouseID != nil {
			if c.WarehouseID == nil || *c.WarehouseID != *filter.WarehouseID {
				continue
			}
		}

		result = append(result, copyContainer(c))
	}

	return result, nil
}

func (r *InMemoryRepo) Update(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[c.ID]; !ok {
		return ErrNotFound
	}

	r.data[c.ID] = copyContainer(c)
	return nil
}
