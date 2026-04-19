package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ljubushkin/container-management-service/internal/apperror"
	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type Service struct {
	repo          repository.Repository
	typeRepo      repository.ContainerTypeRepository
	warehouseRepo repository.WarehouseRepository
}

func NewService(
	repo repository.Repository,
	typeRepo repository.ContainerTypeRepository,
	warehouseRepo repository.WarehouseRepository,
) *Service {
	return &Service{
		repo:          repo,
		typeRepo:      typeRepo,
		warehouseRepo: warehouseRepo,
	}
}

func (s *Service) CreateContainer(typeCode string) (*domain.Container, error) {
	if typeCode == "" {
		return nil, apperror.New(
			apperror.CodeInvalidType,
			"type is required",
			nil,
		)
	}

	_, err := s.typeRepo.GetByCode(typeCode)
	if err != nil {
		return nil, apperror.New(
			apperror.CodeInvalidType,
			"invalid type",
			err,
		)
	}

	c := &domain.Container{
		ID:        uuid.New().String(),
		TypeCode:  typeCode,
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(c); err != nil {
		return nil, apperror.New(
			apperror.CodeInternal,
			"failed to create container",
			err,
		)
	}

	return c, nil
}

func (s *Service) CreateBatch(typeCode string, count int) ([]*domain.Container, error) {
	if count <= 0 {
		return nil, apperror.New(
			apperror.CodeInvalidType,
			"count must be greater than 0",
			nil,
		)
	}

	if typeCode == "" {
		return nil, apperror.New(
			apperror.CodeInvalidType,
			"type is required",
			nil,
		)
	}

	_, err := s.typeRepo.GetByCode(typeCode)
	if err != nil {
		return nil, apperror.New(
			apperror.CodeInvalidType,
			"invalid type",
			err,
		)
	}

	now := time.Now()
	result := make([]*domain.Container, 0, count)

	for range count {
		result = append(result, &domain.Container{
			ID:        uuid.New().String(),
			TypeCode:  typeCode,
			Status:    domain.StatusValid,
			CreatedAt: now,
		})
	}

	if err := s.repo.CreateBatch(result); err != nil {
		return nil, apperror.New(
			apperror.CodeInternal,
			"failed to create batch",
			err,
		)
	}

	return result, nil
}

func (s *Service) GetByID(id string) (*domain.Container, error) {
	container, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, apperror.New(
				apperror.CodeNotFound,
				"container not found",
				err,
			)
		}

		return nil, apperror.New(
			apperror.CodeInternal,
			"failed to get container",
			err,
		)
	}

	return container, nil
}

func (s *Service) UpdateStatus(id string, status domain.Status) error {
	if !domain.IsValidStatus(status) {
		return apperror.New(
			apperror.CodeInvalidStatus,
			"invalid status",
			nil,
		)
	}

	c, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(
				apperror.CodeNotFound,
				"container not found",
				err,
			)
		}

		return apperror.New(
			apperror.CodeInternal,
			"failed to get container",
			err,
		)
	}

	c.Status = status

	if err := s.repo.Update(c); err != nil {
		return apperror.New(
			apperror.CodeInternal,
			"failed to update container",
			err,
		)
	}

	return nil
}

func (s *Service) AssignWarehouse(id string, wid string) error {
	_, err := s.warehouseRepo.GetByID(wid)
	if err != nil {
		return apperror.New(
			apperror.CodeInvalidWarehouse,
			"invalid warehouse",
			err,
		)
	}

	c, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return apperror.New(
				apperror.CodeNotFound,
				"container not found",
				err,
			)
		}

		return apperror.New(
			apperror.CodeInternal,
			"failed to get container",
			err,
		)
	}

	c.WarehouseID = &wid

	if err := s.repo.Update(c); err != nil {
		return apperror.New(
			apperror.CodeInternal,
			"failed to update container",
			err,
		)
	}

	return nil
}

func (s *Service) List(filter domain.ContainerFilter) ([]*domain.Container, error) {
	// базовая валидация pagination
	if filter.Limit < 0 || filter.Offset < 0 {
		return nil, apperror.New(
			apperror.CodeInvalidPagination,
			"invalid pagination params",
			nil,
		)
	}

	// можно задать дефолты (production-паттерн)
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	containers, err := s.repo.List(filter)
	if err != nil {
		return nil, apperror.New(
			apperror.CodeInternal,
			"failed to list containers",
			err,
		)
	}

	return containers, nil
}

func (s *Service) ListWarehouses() ([]*domain.Warehouse, error) {
	warehouses, err := s.warehouseRepo.List()
	if err != nil {
		return nil, apperror.New(
			apperror.CodeInternal,
			"failed to list warehouses",
			err,
		)
	}
	return warehouses, nil
}

func (s *Service) ListContainerTypes() ([]*domain.ContainerType, error) {
	containerTypes, err := s.typeRepo.List()
	if err != nil {
		return nil, apperror.New(
			apperror.CodeInternal,
			"failed to list containerTypes",
			err,
		)
	}
	return containerTypes, nil
}
