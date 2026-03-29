package inmemory

import (
	"sort"
	"sync"

	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type ContainerRepo struct {
	mu   sync.RWMutex
	data map[string]*domain.Container
}

func NewContainerRepo() *ContainerRepo {
	return &ContainerRepo{
		data: make(map[string]*domain.Container),
	}
}

func (r *ContainerRepo) Create(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[c.ID] = cloneContainer(c)
	return nil
}

func (r *ContainerRepo) CreateBatch(containers []*domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, c := range containers {
		r.data[c.ID] = cloneContainer(c)
	}
	return nil
}

func (r *ContainerRepo) GetByID(id string) (*domain.Container, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return cloneContainer(c), nil
}

func (r *ContainerRepo) List(filter domain.ContainerFilter) ([]*domain.Container, error) {
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

func (r *ContainerRepo) Update(c *domain.Container) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[c.ID]; !ok {
		return repository.ErrNotFound
	}

	r.data[c.ID] = cloneContainer(c)
	return nil
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
