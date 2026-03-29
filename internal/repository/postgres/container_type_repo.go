package postgres

import (
	"database/sql"
	"errors"

	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type ContainerTypeRepo struct {
	db *sql.DB
}

func NewContainerTypeRepo(db *sql.DB) *ContainerTypeRepo {
	return &ContainerTypeRepo{db: db}
}

func (r *ContainerTypeRepo) GetByCode(code string) (*domain.ContainerType, error) {
	const query = `
		SELECT code, name
		FROM container_types
		WHERE code = $1
	`

	var t domain.ContainerType
	err := r.db.QueryRow(query, code).Scan(&t.Code, &t.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &t, nil
}

func (r *ContainerTypeRepo) List() ([]*domain.ContainerType, error) {
	const query = `
		SELECT code, name
		FROM container_types
		ORDER BY code
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*domain.ContainerType, 0)
	for rows.Next() {
		var t domain.ContainerType
		if err := rows.Scan(&t.Code, &t.Name); err != nil {
			return nil, err
		}
		result = append(result, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
