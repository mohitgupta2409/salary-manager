package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

func (r *jobTitleRepo) List(ctx context.Context, req model.JobTitleListRequest) (*model.JobTitleListResult, error) {
	page := req.Pagination.Normalized()

	var conds []string
	var args []interface{}
	if !req.IncludeInactive {
		conds = append(conds, "jt.is_active = 1")
	}
	if req.DepartmentID > 0 {
		conds = append(conds, "jt.department_id = ?")
		args = append(args, req.DepartmentID)
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM job_titles jt` + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count job titles: %w", err)
	}

	query := `
		SELECT jt.id, jt.name, jt.department_id, d.name AS department_name,
		       jt.is_active, jt.created_at, jt.updated_at
		FROM job_titles jt
		JOIN departments d ON jt.department_id = d.id` + where + `
		ORDER BY d.name, jt.name`
	query, args = applyPagination(query, args, page)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list job titles: %w", err)
	}
	defer rows.Close()

	out := []model.JobTitle{}
	for rows.Next() {
		var jt model.JobTitle
		if err := rows.Scan(&jt.ID, &jt.Name, &jt.DepartmentID, &jt.Department,
			&jt.IsActive, &jt.CreatedAt, &jt.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan job title: %w", err)
		}
		out = append(out, jt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return &model.JobTitleListResult{
		JobTitles: out,
		Total:     total,
		Limit:     page.Limit,
		Offset:    page.Offset,
	}, nil
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
