package inmemory

import (
	"errors"
	"testing"

	"github.com/ljubushkin/container-management-service/internal/repository"
)

func TestWarehouseRepo_GetByID_Success(t *testing.T) {
	repo := NewWarehouseRepo()

	got, err := repo.GetByID("w1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected warehouse, got nil")
	}
	if got.ID != "w1" {
		t.Fatalf("expected id w1, got %s", got.ID)
	}
	if got.Name != "Main" {
		t.Fatalf("expected name Main, got %s", got.Name)
	}
}

func TestWarehouseRepo_GetByID_NotFound(t *testing.T) {
	repo := NewWarehouseRepo()

	_, err := repo.GetByID("missing")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestWarehouseRepo_List(t *testing.T) {
	repo := NewWarehouseRepo()

	items, err := repo.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 warehouses, got %d", len(items))
	}

	foundW1 := false
	foundW2 := false

	for _, item := range items {
		switch item.ID {
		case "w1":
			foundW1 = true
			if item.Name != "Main" {
				t.Fatalf("expected w1 name Main, got %s", item.Name)
			}
		case "w2":
			foundW2 = true
			if item.Name != "Reserve" {
				t.Fatalf("expected w2 name Reserve, got %s", item.Name)
			}
		}
	}

	if !foundW1 {
		t.Fatal("expected w1 in list")
	}
	if !foundW2 {
		t.Fatal("expected w2 in list")
	}
}
