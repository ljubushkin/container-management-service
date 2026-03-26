package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

var ErrInvalidStatus = errors.New("invalid status")
var ErrInvalidType = errors.New("type is required")

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

var now = time.Now()

func (s *Service) CreateContainer(typeCode string) (*domain.Container, error) {
	if typeCode == "" {
		return nil, errors.New("type is required")
	}

	// 🔥 проверка справочника
	_, err := s.typeRepo.GetByCode(typeCode)
	if err != nil {
		return nil, ErrInvalidType
	}

	c := &domain.Container{
		ID:        uuid.New().String(),
		TypeCode:  typeCode,
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Service) CreateBatch(typeCode string, count int) ([]*domain.Container, error) {
	if count <= 0 {
		return nil, errors.New("count must be greater than 0")
	}

	_, err := s.typeRepo.GetByCode(typeCode)
	if err != nil {
		return nil, errors.New("invalid container type")
	}

	result := make([]*domain.Container, 0, count)

	for range count {
		result = append(result, &domain.Container{
			ID:        uuid.New().String(),
			TypeCode:  typeCode,
			Status:    domain.StatusValid,
			CreatedAt: time.Now(),
		})
	}

	if err := s.repo.CreateBatch(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) GetByID(id string) (*domain.Container, error) {
	return s.repo.GetByID(id)
}

func (s *Service) UpdateStatus(id string, status domain.Status) error {
	if !domain.IsValidStatus(status) {
		return ErrInvalidStatus
	}

	c, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	c.Status = status
	return s.repo.Update(c)
}

func (s *Service) AssignWarehouse(id string, wid string) error {
	_, err := s.warehouseRepo.GetByID(wid)
	if err != nil {
		return errors.New("invalid warehouse")
	}

	c, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	widCopy := wid
	c.WarehouseID = &widCopy
	return s.repo.Update(c)
}
