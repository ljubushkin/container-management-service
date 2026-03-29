package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/repository"
)

type ContainerRepo struct {
	db *sql.DB
}

func NewContainerRepo(db *sql.DB) *ContainerRepo {
	return &ContainerRepo{db: db}
}

func (r *ContainerRepo) Create(c *domain.Container) error {
	const query = `
		INSERT INTO containers (id, type_code, warehouse_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		query,
		c.ID,
		c.TypeCode,
		c.WarehouseID,
		c.Status,
		c.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *ContainerRepo) GetByID(id string) (*domain.Container, error) {
	const query = `
		SELECT id, type_code, warehouse_id, status, created_at
		FROM containers
		WHERE id = $1
	`

	var c domain.Container
	var warehouseID sql.NullString
	var status string

	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.TypeCode,
		&warehouseID,
		&status,
		&c.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	c.Status = domain.Status(status)
	if warehouseID.Valid {
		c.WarehouseID = &warehouseID.String
	}

	return &c, nil
}

func (r *ContainerRepo) Update(c *domain.Container) error {
	const query = `
		UPDATE containers
		SET type_code = $2,
		    warehouse_id = $3,
		    status = $4
		WHERE id = $1
	`

	res, err := r.db.Exec(
		query,
		c.ID,
		c.TypeCode,
		c.WarehouseID,
		c.Status,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *ContainerRepo) CreateBatch(containers []*domain.Container) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO containers (id, type_code, warehouse_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, c := range containers {
		_, err := tx.Exec(
			query,
			c.ID,
			c.TypeCode,
			c.WarehouseID,
			c.Status,
			c.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *ContainerRepo) List(filter domain.ContainerFilter) ([]*domain.Container, error) {
	baseQuery := `
		SELECT id, type_code, warehouse_id, status, created_at
		FROM containers
	`

	var (
		where  []string
		args   []any
		argPos = 1
	)

	if filter.TypeCode != nil {
		where = append(where, fmt.Sprintf("type_code = $%d", argPos))
		args = append(args, *filter.TypeCode)
		argPos++
	}

	if filter.WarehouseID != nil {
		where = append(where, fmt.Sprintf("warehouse_id = $%d", argPos))
		args = append(args, *filter.WarehouseID)
		argPos++
	}

	if filter.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", argPos))
		args = append(args, string(*filter.Status))
		argPos++
	}

	query := baseQuery

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY created_at ASC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*domain.Container, 0)

	for rows.Next() {
		var c domain.Container
		var warehouseID sql.NullString
		var status string

		if err := rows.Scan(
			&c.ID,
			&c.TypeCode,
			&warehouseID,
			&status,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		c.Status = domain.Status(status)

		if warehouseID.Valid {
			wid := warehouseID.String
			c.WarehouseID = &wid
		}

		result = append(result, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
