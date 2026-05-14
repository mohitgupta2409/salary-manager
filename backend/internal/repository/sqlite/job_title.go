package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type jobTitleRepo struct {
	db *sql.DB
}

func NewJobTitleRepository(db *sql.DB) repository.JobTitleRepository {
	return &jobTitleRepo{db: db}
}

func (r *jobTitleRepo) Create(ctx context.Context, jt *model.JobTitle) error {
	now := time.Now().UTC()
	jt.CreatedAt = now
	jt.UpdatedAt = now
	jt.IsActive = true

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO job_titles (name, department_id, is_active, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?)`,
		jt.Name, jt.DepartmentID, jt.CreatedAt, jt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert job title: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	jt.ID = id
	return nil
}

func (r *jobTitleRepo) List(ctx context.Context, departmentID int64, includeInactive bool) ([]model.JobTitle, error) {
	query := `
		SELECT jt.id, jt.name, jt.department_id, d.name AS department_name,
		       jt.is_active, jt.created_at, jt.updated_at
		FROM job_titles jt
		JOIN departments d ON jt.department_id = d.id`

	var conds []string
	var args []interface{}
	if !includeInactive {
		conds = append(conds, "jt.is_active = 1")
	}
	if departmentID > 0 {
		conds = append(conds, "jt.department_id = ?")
		args = append(args, departmentID)
	}
	if len(conds) > 0 {
		query += " WHERE " + joinConds(conds)
	}
	query += " ORDER BY d.name, jt.name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list job titles: %w", err)
	}
	defer rows.Close()

	var out []model.JobTitle
	for rows.Next() {
		var jt model.JobTitle
		if err := rows.Scan(&jt.ID, &jt.Name, &jt.DepartmentID, &jt.Department,
			&jt.IsActive, &jt.CreatedAt, &jt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan job title: %w", err)
		}
		out = append(out, jt)
	}
	return out, rows.Err()
}

func (r *jobTitleRepo) GetByID(ctx context.Context, id int64) (*model.JobTitle, error) {
	jt := &model.JobTitle{}
	err := r.db.QueryRowContext(ctx, `
		SELECT jt.id, jt.name, jt.department_id, d.name AS department_name,
		       jt.is_active, jt.created_at, jt.updated_at
		FROM job_titles jt
		JOIN departments d ON jt.department_id = d.id
		WHERE jt.id = ?`, id,
	).Scan(&jt.ID, &jt.Name, &jt.DepartmentID, &jt.Department,
		&jt.IsActive, &jt.CreatedAt, &jt.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job title: %w", err)
	}
	return jt, nil
}

func joinConds(conds []string) string {
	out := ""
	for i, c := range conds {
		if i > 0 {
			out += " AND "
		}
		out += c
	}
	return out
}
