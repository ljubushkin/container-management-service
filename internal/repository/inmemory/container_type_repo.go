package inmemory

import (
	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type ContainerTypeRepo struct {
	data map[string]*domain.ContainerType
}

func NewContainerTypeRepo() *ContainerTypeRepo {
	return &ContainerTypeRepo{
		data: map[string]*domain.ContainerType{
			"EURO_PALLET": {Code: "EURO_PALLET", Name: "Euro pallet"},
			"BOX":         {Code: "BOX", Name: "Box"},
		},
	}
}

func (r *ContainerTypeRepo) GetByCode(code string) (*domain.ContainerType, error) {
	t, ok := r.data[code]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return t, nil
}

func (r *ContainerTypeRepo) List() ([]*domain.ContainerType, error) {
	result := make([]*domain.ContainerType, 0, len(r.data))
	for _, t := range r.data {
		result = append(result, t)
	}
	return result, nil
}
