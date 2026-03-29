package postgres

import (
	"database/sql"
	"errors"

	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type WarehouseRepo struct {
	db *sql.DB
}

func NewWarehouseRepo(db *sql.DB) *WarehouseRepo {
	return &WarehouseRepo{db: db}
}

func (r *WarehouseRepo) GetByID(id string) (*domain.Warehouse, error) {
	const query = `
		SELECT id, name
		FROM warehouses
		WHERE id = $1
	`

	var w domain.Warehouse
	err := r.db.QueryRow(query, id).Scan(&w.ID, &w.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &w, nil
}

func (r *WarehouseRepo) List() ([]*domain.Warehouse, error) {
	const query = `
		SELECT id, name
		FROM warehouses
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*domain.Warehouse, 0)
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(&w.ID, &w.Name); err != nil {
			return nil, err
		}
		result = append(result, &w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
