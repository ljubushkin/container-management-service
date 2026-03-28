package repository

import (
	"errors"
	"sort"
	"sync"

	"github.com/ljubushkin/container-management-service/internal/domain"
)

var ErrNotFound = errors.New("not found")

// ===== Interfaces =====

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

// ===== InMemory Container Repo =====

type InMemoryRepo struct {
	mu   sync.RWMutex
	data map[string]*domain.Container
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{
		data: make(map[string]*domain.Container),
	}
}

func (r *InMemoryRepo) Create(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[c.ID] = cloneContainer(c)
	return nil
}

func (r *InMemoryRepo) CreateBatch(containers []*domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range containers {
		r.data[c.ID] = cloneContainer(c)
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

	return cloneContainer(c), nil
}

func (r *InMemoryRepo) List(filter domain.ContainerFilter) ([]*domain.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Container, 0)

	for _, c := range r.data {
		if filter.TypeCode != nil && c.TypeCode != *filter.TypeCode {
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

		result = append(result, cloneContainer(c))
	}

	// стабильный порядок (важно для pagination)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	// pagination
	start := filter.Offset
	if start > len(result) {
		return []*domain.Container{}, nil
	}

	end := start + filter.Limit
	if end > len(result) {
		end = len(result)
	}

	return result[start:end], nil
}

func (r *InMemoryRepo) Update(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[c.ID]; !ok {
		return ErrNotFound
	}

	r.data[c.ID] = cloneContainer(c)
	return nil
}

// ===== InMemory Type Repo =====

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

// ===== InMemory Warehouse Repo =====

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

// ===== Helpers =====

func cloneContainer(c *domain.Container) *domain.Container {
	if c == nil {
		return nil
	}

	clone := *c

	if c.WarehouseID != nil {
		wid := *c.WarehouseID
		clone.WarehouseID = &wid
	}

	return &clone
}
