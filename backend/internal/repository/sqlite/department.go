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

func (r *departmentRepo) List(ctx context.Context, req model.DepartmentListRequest) (*model.DepartmentListResult, error) {
	page := req.Pagination.Normalized()

	where := ""
	var args []interface{}
	if !req.IncludeInactive {
		where = " WHERE is_active = 1"
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM departments`+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count departments: %w", err)
	}

	query := `SELECT id, name, is_active, created_at, updated_at
	          FROM departments` + where + ` ORDER BY name`
	query, args = applyPagination(query, args, page)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}
	defer rows.Close()

	out := []model.Department{}
	for rows.Next() {
		var d model.Department
		if err := rows.Scan(&d.ID, &d.Name, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return &model.DepartmentListResult{
		Departments: out,
		Total:       total,
		Limit:       page.Limit,
		Offset:      page.Offset,
	}, nil
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
