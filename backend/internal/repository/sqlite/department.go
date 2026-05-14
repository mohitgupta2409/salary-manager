package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type departmentRepo struct {
	db *sql.DB
}

func NewDepartmentRepository(db *sql.DB) repository.DepartmentRepository {
	return &departmentRepo{db: db}
}

func (r *departmentRepo) Create(ctx context.Context, d *model.Department) error {
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now
	d.IsActive = true

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO departments (name, is_active, created_at, updated_at)
		VALUES (?, 1, ?, ?)`,
		d.Name, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert department: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	d.ID = id
	return nil
}

func (r *departmentRepo) List(ctx context.Context, includeInactive bool) ([]model.Department, error) {
	query := `SELECT id, name, is_active, created_at, updated_at FROM departments`
	if !includeInactive {
		query += ` WHERE is_active = 1`
	}
	query += ` ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()

	var out []model.Department
	for rows.Next() {
		var d model.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *departmentRepo) GetByID(ctx context.Context, id int64) (*model.Department, error) {
	d := &model.Department{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, is_active, created_at, updated_at
		FROM departments WHERE id = ?`, id,
	).Scan(&d.ID, &d.Name, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}
	return d, nil
}
