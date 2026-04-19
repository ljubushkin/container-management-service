package inmemory

import (
	"errors"
	"testing"

	"github.com/ljubushkin/container-management-service/internal/repository"
)

func TestContainerTypeRepo_GetByCode_Success(t *testing.T) {
	repo := NewContainerTypeRepo()

	got, err := repo.GetByCode("BOX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("expected container type, got nil")
	}
	if got.Code != "BOX" {
		t.Fatalf("expected code BOX, got %s", got.Code)
	}
	if got.Name == "" {
		t.Fatal("expected non-empty name")
	}
}

func TestContainerTypeRepo_GetByCode_NotFound(t *testing.T) {
	repo := NewContainerTypeRepo()

	_, err := repo.GetByCode("UNKNOWN")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestContainerTypeRepo_List(t *testing.T) {
	repo := NewContainerTypeRepo()

	items, err := repo.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 container types, got %d", len(items))
	}

	foundBox := false
	foundEuro := false

	for _, item := range items {
		switch item.Code {
		case "BOX":
			foundBox = true
		case "EURO_PALLET":
			foundEuro = true
		}
	}

	if !foundBox {
		t.Fatal("expected BOX in list")
	}
	if !foundEuro {
		t.Fatal("expected EURO_PALLET in list")
	}
}
