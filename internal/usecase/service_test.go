package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/ljubushkin/container-management-service/internal/apperror"
	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type repoMock struct {
	createFn      func(c *domain.Container) error
	createBatchFn func(containers []*domain.Container) error
	getByIDFn     func(id string) (*domain.Container, error)
	listFn        func(filter domain.ContainerFilter) ([]*domain.Container, error)
	updateFn      func(c *domain.Container) error
}

func (m *repoMock) Create(c *domain.Container) error {
	if m.createFn != nil {
		return m.createFn(c)
	}
	return nil
}

func (m *repoMock) CreateBatch(containers []*domain.Container) error {
	if m.createBatchFn != nil {
		return m.createBatchFn(containers)
	}
	return nil
}

func (m *repoMock) GetByID(id string) (*domain.Container, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}

func (m *repoMock) List(filter domain.ContainerFilter) ([]*domain.Container, error) {
	if m.listFn != nil {
		return m.listFn(filter)
	}
	return nil, nil
}

func (m *repoMock) Update(c *domain.Container) error {
	if m.updateFn != nil {
		return m.updateFn(c)
	}
	return nil
}

type typeRepoMock struct {
	getByCodeFn func(code string) (*domain.ContainerType, error)
	listFn      func() ([]*domain.ContainerType, error)
}

func (m *typeRepoMock) GetByCode(code string) (*domain.ContainerType, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(code)
	}
	return nil, nil
}

func (m *typeRepoMock) List() ([]*domain.ContainerType, error) {
	if m.listFn != nil {
		return m.listFn()
	}
	return nil, nil
}

type warehouseRepoMock struct {
	getByIDFn func(id string) (*domain.Warehouse, error)
	listFn    func() ([]*domain.Warehouse, error)
}

func (m *warehouseRepoMock) GetByID(id string) (*domain.Warehouse, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}

func (m *warehouseRepoMock) List() ([]*domain.Warehouse, error) {
	if m.listFn != nil {
		return m.listFn()
	}
	return nil, nil
}

func assertAppErrorCode(t *testing.T, err error, expected apperror.Code) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != expected {
		t.Fatalf("expected code %q, got %q", expected, appErr.Code)
	}
}

func TestService_CreateContainer_Success(t *testing.T) {
	repo := &repoMock{
		createFn: func(c *domain.Container) error {
			if c.ID == "" {
				t.Fatal("expected generated ID")
			}
			if c.TypeCode != "BOX" {
				t.Fatalf("expected type BOX, got %s", c.TypeCode)
			}
			if c.Status != domain.StatusValid {
				t.Fatalf("expected status valid, got %s", c.Status)
			}
			if c.CreatedAt.IsZero() {
				t.Fatal("expected CreatedAt to be set")
			}
			return nil
		},
	}

	typeRepo := &typeRepoMock{
		getByCodeFn: func(code string) (*domain.ContainerType, error) {
			if code != "BOX" {
				t.Fatalf("expected code BOX, got %s", code)
			}
			return &domain.ContainerType{Code: "BOX", Name: "Box"}, nil
		},
	}

	svc := NewService(repo, typeRepo, &warehouseRepoMock{})

	c, err := svc.CreateContainer("BOX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c == nil {
		t.Fatal("expected container, got nil")
	}
	if c.TypeCode != "BOX" {
		t.Fatalf("expected type BOX, got %s", c.TypeCode)
	}
	if c.Status != domain.StatusValid {
		t.Fatalf("expected status valid, got %s", c.Status)
	}
}

func TestService_CreateContainer_EmptyType(t *testing.T) {
	svc := NewService(&repoMock{}, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.CreateContainer("")
	assertAppErrorCode(t, err, apperror.CodeInvalidType)
}

func TestService_CreateContainer_InvalidType(t *testing.T) {
	typeRepo := &typeRepoMock{
		getByCodeFn: func(code string) (*domain.ContainerType, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewService(&repoMock{}, typeRepo, &warehouseRepoMock{})

	_, err := svc.CreateContainer("BAD")
	assertAppErrorCode(t, err, apperror.CodeInvalidType)
}

func TestService_CreateContainer_CreateFailed(t *testing.T) {
	repo := &repoMock{
		createFn: func(c *domain.Container) error {
			return errors.New("create failed")
		},
	}
	typeRepo := &typeRepoMock{
		getByCodeFn: func(code string) (*domain.ContainerType, error) {
			return &domain.ContainerType{Code: code, Name: "Type"}, nil
		},
	}

	svc := NewService(repo, typeRepo, &warehouseRepoMock{})

	_, err := svc.CreateContainer("BOX")
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_CreateBatch_Success(t *testing.T) {
	repo := &repoMock{
		createBatchFn: func(containers []*domain.Container) error {
			if len(containers) != 3 {
				t.Fatalf("expected 3 containers, got %d", len(containers))
			}

			var firstCreatedAt time.Time
			for i, c := range containers {
				if c.ID == "" {
					t.Fatalf("container %d: expected generated ID", i)
				}
				if c.TypeCode != "BOX" {
					t.Fatalf("container %d: expected type BOX, got %s", i, c.TypeCode)
				}
				if c.Status != domain.StatusValid {
					t.Fatalf("container %d: expected status valid, got %s", i, c.Status)
				}
				if c.CreatedAt.IsZero() {
					t.Fatalf("container %d: expected CreatedAt", i)
				}
				if i == 0 {
					firstCreatedAt = c.CreatedAt
				} else if !c.CreatedAt.Equal(firstCreatedAt) {
					t.Fatalf("container %d: expected same CreatedAt for batch", i)
				}
			}
			return nil
		},
	}
	typeRepo := &typeRepoMock{
		getByCodeFn: func(code string) (*domain.ContainerType, error) {
			return &domain.ContainerType{Code: code, Name: "Type"}, nil
		},
	}

	svc := NewService(repo, typeRepo, &warehouseRepoMock{})

	containers, err := svc.CreateBatch("BOX", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(containers) != 3 {
		t.Fatalf("expected 3 containers, got %d", len(containers))
	}
}

func TestService_CreateBatch_InvalidCount(t *testing.T) {
	svc := NewService(&repoMock{}, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.CreateBatch("BOX", 0)
	assertAppErrorCode(t, err, apperror.CodeInvalidType)
}

func TestService_CreateBatch_EmptyType(t *testing.T) {
	svc := NewService(&repoMock{}, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.CreateBatch("", 2)
	assertAppErrorCode(t, err, apperror.CodeInvalidType)
}

func TestService_CreateBatch_InvalidType(t *testing.T) {
	typeRepo := &typeRepoMock{
		getByCodeFn: func(code string) (*domain.ContainerType, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewService(&repoMock{}, typeRepo, &warehouseRepoMock{})

	_, err := svc.CreateBatch("BAD", 2)
	assertAppErrorCode(t, err, apperror.CodeInvalidType)
}

func TestService_CreateBatch_CreateFailed(t *testing.T) {
	repo := &repoMock{
		createBatchFn: func(containers []*domain.Container) error {
			return errors.New("batch failed")
		},
	}
	typeRepo := &typeRepoMock{
		getByCodeFn: func(code string) (*domain.ContainerType, error) {
			return &domain.ContainerType{Code: code, Name: "Type"}, nil
		},
	}

	svc := NewService(repo, typeRepo, &warehouseRepoMock{})

	_, err := svc.CreateBatch("BOX", 2)
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_GetByID_Success(t *testing.T) {
	expected := &domain.Container{
		ID:        "c1",
		TypeCode:  "BOX",
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	}

	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			if id != "c1" {
				t.Fatalf("expected id c1, got %s", id)
			}
			return expected, nil
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	got, err := svc.GetByID("c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != expected {
		t.Fatalf("expected same container pointer")
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.GetByID("missing")
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestService_GetByID_InternalError(t *testing.T) {
	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return nil, errors.New("unexpected")
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.GetByID("c1")
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_UpdateStatus_Success(t *testing.T) {
	existing := &domain.Container{
		ID:        "c1",
		TypeCode:  "BOX",
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	}

	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return existing, nil
		},
		updateFn: func(c *domain.Container) error {
			if c.Status != domain.StatusDefect {
				t.Fatalf("expected updated status defect, got %s", c.Status)
			}
			return nil
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	err := svc.UpdateStatus("c1", domain.StatusDefect)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_UpdateStatus_InvalidStatus(t *testing.T) {
	svc := NewService(&repoMock{}, &typeRepoMock{}, &warehouseRepoMock{})

	err := svc.UpdateStatus("c1", domain.Status("bad_status"))
	assertAppErrorCode(t, err, apperror.CodeInvalidStatus)
}

func TestService_UpdateStatus_NotFound(t *testing.T) {
	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	err := svc.UpdateStatus("missing", domain.StatusValid)
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestService_UpdateStatus_UpdateFailed(t *testing.T) {
	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return &domain.Container{
				ID:        id,
				TypeCode:  "BOX",
				Status:    domain.StatusValid,
				CreatedAt: time.Now(),
			}, nil
		},
		updateFn: func(c *domain.Container) error {
			return errors.New("update failed")
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	err := svc.UpdateStatus("c1", domain.StatusDefect)
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_AssignWarehouse_Success(t *testing.T) {
	existing := &domain.Container{
		ID:        "c1",
		TypeCode:  "BOX",
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	}

	warehouseRepo := &warehouseRepoMock{
		getByIDFn: func(id string) (*domain.Warehouse, error) {
			if id != "w1" {
				t.Fatalf("expected warehouse w1, got %s", id)
			}
			return &domain.Warehouse{ID: "w1", Name: "Main"}, nil
		},
	}

	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return existing, nil
		},
		updateFn: func(c *domain.Container) error {
			if c.WarehouseID == nil {
				t.Fatal("expected warehouse id to be set")
			}
			if *c.WarehouseID != "w1" {
				t.Fatalf("expected warehouse w1, got %s", *c.WarehouseID)
			}
			return nil
		},
	}

	svc := NewService(repo, &typeRepoMock{}, warehouseRepo)

	err := svc.AssignWarehouse("c1", "w1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_AssignWarehouse_InvalidWarehouse(t *testing.T) {
	warehouseRepo := &warehouseRepoMock{
		getByIDFn: func(id string) (*domain.Warehouse, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewService(&repoMock{}, &typeRepoMock{}, warehouseRepo)

	err := svc.AssignWarehouse("c1", "bad")
	assertAppErrorCode(t, err, apperror.CodeInvalidWarehouse)
}

func TestService_AssignWarehouse_ContainerNotFound(t *testing.T) {
	warehouseRepo := &warehouseRepoMock{
		getByIDFn: func(id string) (*domain.Warehouse, error) {
			return &domain.Warehouse{ID: "w1", Name: "Main"}, nil
		},
	}
	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return nil, repository.ErrNotFound
		},
	}

	svc := NewService(repo, &typeRepoMock{}, warehouseRepo)

	err := svc.AssignWarehouse("missing", "w1")
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestService_AssignWarehouse_UpdateFailed(t *testing.T) {
	warehouseRepo := &warehouseRepoMock{
		getByIDFn: func(id string) (*domain.Warehouse, error) {
			return &domain.Warehouse{ID: "w1", Name: "Main"}, nil
		},
	}
	repo := &repoMock{
		getByIDFn: func(id string) (*domain.Container, error) {
			return &domain.Container{
				ID:        id,
				TypeCode:  "BOX",
				Status:    domain.StatusValid,
				CreatedAt: time.Now(),
			}, nil
		},
		updateFn: func(c *domain.Container) error {
			return errors.New("update failed")
		},
	}

	svc := NewService(repo, &typeRepoMock{}, warehouseRepo)

	err := svc.AssignWarehouse("c1", "w1")
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_List_SuccessWithDefaultLimit(t *testing.T) {
	expected := []*domain.Container{
		{
			ID:        "c1",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: time.Now(),
		},
	}

	repo := &repoMock{
		listFn: func(filter domain.ContainerFilter) ([]*domain.Container, error) {
			if filter.Limit != 50 {
				t.Fatalf("expected default limit 50, got %d", filter.Limit)
			}
			if filter.Offset != 0 {
				t.Fatalf("expected offset 0, got %d", filter.Offset)
			}
			return expected, nil
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	got, err := svc.List(domain.ContainerFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 container, got %d", len(got))
	}
}

func TestService_List_InvalidPagination(t *testing.T) {
	svc := NewService(&repoMock{}, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.List(domain.ContainerFilter{
		Limit:  -1,
		Offset: 0,
	})
	assertAppErrorCode(t, err, apperror.CodeInvalidPagination)
}

func TestService_List_InternalError(t *testing.T) {
	repo := &repoMock{
		listFn: func(filter domain.ContainerFilter) ([]*domain.Container, error) {
			return nil, errors.New("list failed")
		},
	}

	svc := NewService(repo, &typeRepoMock{}, &warehouseRepoMock{})

	_, err := svc.List(domain.ContainerFilter{Limit: 10})
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_ListWarehouses_Success(t *testing.T) {
	warehouseRepo := &warehouseRepoMock{
		listFn: func() ([]*domain.Warehouse, error) {
			return []*domain.Warehouse{
				{ID: "w1", Name: "Main"},
				{ID: "w2", Name: "Reserve"},
			}, nil
		},
	}

	svc := NewService(&repoMock{}, &typeRepoMock{}, warehouseRepo)

	items, err := svc.ListWarehouses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 warehouses, got %d", len(items))
	}
}

func TestService_ListWarehouses_InternalError(t *testing.T) {
	warehouseRepo := &warehouseRepoMock{
		listFn: func() ([]*domain.Warehouse, error) {
			return nil, errors.New("list warehouses failed")
		},
	}

	svc := NewService(&repoMock{}, &typeRepoMock{}, warehouseRepo)

	_, err := svc.ListWarehouses()
	assertAppErrorCode(t, err, apperror.CodeInternal)
}

func TestService_ListContainerTypes_Success(t *testing.T) {
	typeRepo := &typeRepoMock{
		listFn: func() ([]*domain.ContainerType, error) {
			return []*domain.ContainerType{
				{Code: "BOX", Name: "Box"},
				{Code: "EURO_PALLET", Name: "Euro pallet"},
			}, nil
		},
	}

	svc := NewService(&repoMock{}, typeRepo, &warehouseRepoMock{})

	items, err := svc.ListContainerTypes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 types, got %d", len(items))
	}
}

func TestService_ListContainerTypes_InternalError(t *testing.T) {
	typeRepo := &typeRepoMock{
		listFn: func() ([]*domain.ContainerType, error) {
			return nil, errors.New("list types failed")
		},
	}

	svc := NewService(&repoMock{}, typeRepo, &warehouseRepoMock{})

	_, err := svc.ListContainerTypes()
	assertAppErrorCode(t, err, apperror.CodeInternal)
}
