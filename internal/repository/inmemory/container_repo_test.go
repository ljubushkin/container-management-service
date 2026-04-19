package inmemory

import (
	"errors"
	"testing"
	"time"

	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

func TestContainerRepo_CreateAndGetByID(t *testing.T) {
	repo := NewContainerRepo()

	now := time.Now()
	input := &domain.Container{
		ID:        "c1",
		TypeCode:  "BOX",
		Status:    domain.StatusValid,
		CreatedAt: now,
	}

	if err := repo.Create(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID("c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != input.ID {
		t.Fatalf("expected id %s, got %s", input.ID, got.ID)
	}
	if got.TypeCode != input.TypeCode {
		t.Fatalf("expected type %s, got %s", input.TypeCode, got.TypeCode)
	}
	if got.Status != input.Status {
		t.Fatalf("expected status %s, got %s", input.Status, got.Status)
	}
	if !got.CreatedAt.Equal(input.CreatedAt) {
		t.Fatalf("expected createdAt %v, got %v", input.CreatedAt, got.CreatedAt)
	}
}

func TestContainerRepo_GetByID_NotFound(t *testing.T) {
	repo := NewContainerRepo()

	_, err := repo.GetByID("missing")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContainerRepo_Update_Success(t *testing.T) {
	repo := NewContainerRepo()

	input := &domain.Container{
		ID:        "c1",
		TypeCode:  "BOX",
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	}

	if err := repo.Create(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	warehouseID := "w1"
	updated := &domain.Container{
		ID:          "c1",
		TypeCode:    "BOX",
		Status:      domain.StatusDefect,
		WarehouseID: &warehouseID,
		CreatedAt:   input.CreatedAt,
	}

	if err := repo.Update(updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID("c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Status != domain.StatusDefect {
		t.Fatalf("expected status defect, got %s", got.Status)
	}
	if got.WarehouseID == nil {
		t.Fatal("expected warehouse id to be set")
	}
	if *got.WarehouseID != "w1" {
		t.Fatalf("expected warehouse w1, got %s", *got.WarehouseID)
	}
}

func TestContainerRepo_Update_NotFound(t *testing.T) {
	repo := NewContainerRepo()

	err := repo.Update(&domain.Container{
		ID:        "missing",
		TypeCode:  "BOX",
		Status:    domain.StatusValid,
		CreatedAt: time.Now(),
	})

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContainerRepo_CreateBatch(t *testing.T) {
	repo := NewContainerRepo()

	now := time.Now()
	items := []*domain.Container{
		{
			ID:        "c1",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: now,
		},
		{
			ID:        "c2",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: now,
		},
		{
			ID:        "c3",
			TypeCode:  "EURO_PALLET",
			Status:    domain.StatusDefect,
			CreatedAt: now,
		},
	}

	if err := repo.CreateBatch(items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, id := range []string{"c1", "c2", "c3"} {
		_, err := repo.GetByID(id)
		if err != nil {
			t.Fatalf("expected container %s to exist, got error: %v", id, err)
		}
	}
}

func TestContainerRepo_List_FilterByType(t *testing.T) {
	repo := NewContainerRepo()

	seedContainers(t, repo)

	typeCode := "BOX"
	got, err := repo.List(domain.ContainerFilter{
		TypeCode: &typeCode,
		Limit:    10,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(got))
	}

	for _, c := range got {
		if c.TypeCode != "BOX" {
			t.Fatalf("expected BOX, got %s", c.TypeCode)
		}
	}
}

func TestContainerRepo_List_FilterByStatus(t *testing.T) {
	repo := NewContainerRepo()

	seedContainers(t, repo)

	status := domain.StatusDefect
	got, err := repo.List(domain.ContainerFilter{
		Status: &status,
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 container, got %d", len(got))
	}

	if got[0].Status != domain.StatusDefect {
		t.Fatalf("expected defect, got %s", got[0].Status)
	}
}

func TestContainerRepo_List_FilterByWarehouse(t *testing.T) {
	repo := NewContainerRepo()

	seedContainers(t, repo)

	warehouseID := "w1"
	got, err := repo.List(domain.ContainerFilter{
		WarehouseID: &warehouseID,
		Limit:       10,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 container, got %d", len(got))
	}

	if got[0].WarehouseID == nil || *got[0].WarehouseID != "w1" {
		t.Fatalf("expected warehouse w1")
	}
}

func TestContainerRepo_List_Pagination(t *testing.T) {
	repo := NewContainerRepo()

	t1 := time.Now().Add(-3 * time.Hour)
	t2 := time.Now().Add(-2 * time.Hour)
	t3 := time.Now().Add(-1 * time.Hour)

	items := []*domain.Container{
		{
			ID:        "c1",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: t1,
		},
		{
			ID:        "c2",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: t2,
		},
		{
			ID:        "c3",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: t3,
		},
	}

	if err := repo.CreateBatch(items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.List(domain.ContainerFilter{
		Limit:  2,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(got))
	}

	if got[0].ID != "c2" {
		t.Fatalf("expected first result c2, got %s", got[0].ID)
	}
	if got[1].ID != "c3" {
		t.Fatalf("expected second result c3, got %s", got[1].ID)
	}
}

func TestContainerRepo_List_OffsetGreaterThanLength(t *testing.T) {
	repo := NewContainerRepo()

	seedContainers(t, repo)

	got, err := repo.List(domain.ContainerFilter{
		Limit:  10,
		Offset: 100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}

func TestContainerRepo_GetByID_ReturnsCopy(t *testing.T) {
	repo := NewContainerRepo()

	warehouseID := "w1"
	input := &domain.Container{
		ID:          "c1",
		TypeCode:    "BOX",
		Status:      domain.StatusValid,
		WarehouseID: &warehouseID,
		CreatedAt:   time.Now(),
	}

	if err := repo.Create(input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID("c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got.Status = domain.StatusWritten
	*got.WarehouseID = "changed"

	again, err := repo.GetByID("c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if again.Status != domain.StatusValid {
		t.Fatalf("expected original status valid, got %s", again.Status)
	}
	if again.WarehouseID == nil {
		t.Fatal("expected warehouse id to be set")
	}
	if *again.WarehouseID != "w1" {
		t.Fatalf("expected original warehouse w1, got %s", *again.WarehouseID)
	}
}

func TestContainerRepo_List_ReturnsCopies(t *testing.T) {
	repo := NewContainerRepo()

	seedContainers(t, repo)

	got, err := repo.List(domain.ContainerFilter{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("expected non-empty result")
	}

	got[0].Status = domain.StatusWritten

	again, err := repo.List(domain.ContainerFilter{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if again[0].Status == domain.StatusWritten {
		t.Fatal("expected repo data to remain unchanged")
	}
}

func seedContainers(t *testing.T, repo *ContainerRepo) {
	t.Helper()

	w1 := "w1"
	now := time.Now()

	items := []*domain.Container{
		{
			ID:        "c1",
			TypeCode:  "BOX",
			Status:    domain.StatusValid,
			CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			ID:          "c2",
			TypeCode:    "BOX",
			Status:      domain.StatusValid,
			WarehouseID: &w1,
			CreatedAt:   now.Add(-2 * time.Hour),
		},
		{
			ID:        "c3",
			TypeCode:  "EURO_PALLET",
			Status:    domain.StatusDefect,
			CreatedAt: now.Add(-1 * time.Hour),
		},
	}

	if err := repo.CreateBatch(items); err != nil {
		t.Fatalf("unexpected error while seeding: %v", err)
	}
}
