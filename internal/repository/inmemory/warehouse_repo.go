package inmemory

import (
	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type WarehouseRepo struct {
	data map[string]*domain.Warehouse
}

func NewWarehouseRepo() *WarehouseRepo {
	return &WarehouseRepo{
		data: map[string]*domain.Warehouse{
			"w1": {ID: "w1", Name: "Main"},
			"w2": {ID: "w2", Name: "Reserve"},
		},
	}
}

func (r *WarehouseRepo) GetByID(id string) (*domain.Warehouse, error) {
	w, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return w, nil
}

func (r *WarehouseRepo) List() ([]*domain.Warehouse, error) {
	result := make([]*domain.Warehouse, 0, len(r.data))
	for _, w := range r.data {
		result = append(result, w)
	}
	return result, nil
}
