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

type employeeRepo struct {
	db *sql.DB
}

func NewEmployeeRepository(db *sql.DB) repository.EmployeeRepository {
	return &employeeRepo{db: db}
}

// Common SELECT used by all read paths. Joins employee with its referenced
// country, job_title, and department so callers receive denormalized fields
// (country name, currency, job title, department) in a single round-trip.
const employeeSelectClause = `
	SELECT e.id, e.first_name, e.last_name, e.email,
	       e.job_title_id, e.country_id, e.salary, e.address,
	       e.join_date, e.is_active, e.created_at, e.updated_at,
	       jt.name AS job_title, d.name AS department,
	       c.name AS country, c.currency AS currency
	FROM employees e
	JOIN job_titles  jt ON e.job_title_id  = jt.id
	JOIN departments d  ON jt.department_id = d.id
	JOIN countries   c  ON e.country_id    = c.id
`

func scanEmployee(rows interface {
	Scan(dest ...interface{}) error
}) (*model.Employee, error) {
	var e model.Employee
	var address sql.NullString
	err := rows.Scan(
		&e.ID, &e.FirstName, &e.LastName, &e.Email,
		&e.JobTitleID, &e.CountryID, &e.Salary, &address,
		&e.JoinDate, &e.IsActive, &e.CreatedAt, &e.UpdatedAt,
		&e.JobTitle, &e.Department, &e.Country, &e.Currency,
	)
	if err != nil {
		return nil, err
	}
	if address.Valid {
		e.Address = address.String
	}
	return &e, nil
}

func (r *employeeRepo) Create(ctx context.Context, emp *model.Employee) error {
	now := time.Now().UTC()
	emp.CreatedAt = now
	emp.UpdatedAt = now
	emp.IsActive = true

	var address interface{}
	if emp.Address != "" {
		address = emp.Address
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO employees (first_name, last_name, email, job_title_id, country_id,
		                       salary, address, join_date, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`,
		emp.FirstName, emp.LastName, emp.Email, emp.JobTitleID, emp.CountryID,
		emp.Salary, address, emp.JoinDate, emp.CreatedAt, emp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert employee: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	emp.ID = id

	// Hydrate denormalized fields by re-fetching
	fresh, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if fresh != nil {
		*emp = *fresh
	}
	return nil
}

func (r *employeeRepo) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	row := r.db.QueryRowContext(ctx, employeeSelectClause+`WHERE e.id = ?`, id)
	emp, err := scanEmployee(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get employee: %w", err)
	}
	return emp, nil
}

func (r *employeeRepo) Update(ctx context.Context, emp *model.Employee) error {
	emp.UpdatedAt = time.Now().UTC()

	var address interface{}
	if emp.Address != "" {
		address = emp.Address
	}

	res, err := r.db.ExecContext(ctx, `
		UPDATE employees
		SET first_name = ?, last_name = ?, email = ?,
		    job_title_id = ?, country_id = ?, salary = ?, address = ?,
		    join_date = ?, updated_at = ?
		WHERE id = ?`,
		emp.FirstName, emp.LastName, emp.Email,
		emp.JobTitleID, emp.CountryID, emp.Salary, address,
		emp.JoinDate, emp.UpdatedAt, emp.ID,
	)
	if err != nil {
		return fmt.Errorf("update employee: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("employee with id %d not found", emp.ID)
	}

	fresh, err := r.GetByID(ctx, emp.ID)
	if err != nil {
		return err
	}
	if fresh != nil {
		*emp = *fresh
	}
	return nil
}

func (r *employeeRepo) SoftDelete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE employees SET is_active = 0, updated_at = ? WHERE id = ? AND is_active = 1`,
		time.Now().UTC(), id,
	)
	if err != nil {
		return fmt.Errorf("soft delete employee: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("employee with id %d not found", id)
	}
	return nil
}

func (r *employeeRepo) List(ctx context.Context, filter model.EmployeeFilter) (*model.EmployeeListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	conds := []string{"e.is_active = 1"}
	args := []interface{}{}

	if filter.CountryID > 0 {
		conds = append(conds, "e.country_id = ?")
		args = append(args, filter.CountryID)
	}
	if filter.JobTitleID > 0 {
		conds = append(conds, "e.job_title_id = ?")
		args = append(args, filter.JobTitleID)
	}
	if filter.DepartmentID > 0 {
		conds = append(conds, "jt.department_id = ?")
		args = append(args, filter.DepartmentID)
	}
	if filter.Search != "" {
		conds = append(conds, "(e.first_name LIKE ? OR e.last_name LIKE ? OR e.email LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	var total int64
	countQuery := `
		SELECT COUNT(*) FROM employees e
		JOIN job_titles jt ON e.job_title_id = jt.id ` + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count employees: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := employeeSelectClause + where + ` ORDER BY e.id DESC LIMIT ? OFFSET ?`
	dataArgs := append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	var employees []model.Employee
	for rows.Next() {
		emp, err := scanEmployee(rows)
		if err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		employees = append(employees, *emp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return &model.EmployeeListResult{
		Employees: employees,
		Total:     total,
		Page:      filter.Page,
		Limit:     filter.Limit,
	}, nil
}

func (r *employeeRepo) GetSalaryRangeByCountry(ctx context.Context, country string) (*model.SalaryRange, error) {
	sr := &model.SalaryRange{Country: country}

	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MIN(e.salary), 0), COALESCE(MAX(e.salary), 0),
		       COALESCE(AVG(e.salary), 0), COUNT(*)
		FROM employees e
		JOIN countries c ON e.country_id = c.id
		WHERE c.name = ? AND e.is_active = 1`, country,
	).Scan(&sr.Min, &sr.Max, &sr.Average, &sr.Count)
	if err != nil {
		return nil, fmt.Errorf("salary range by country: %w", err)
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(salary, 0) FROM (
			SELECT e.salary,
			       ROW_NUMBER() OVER (ORDER BY e.salary) AS rn,
			       COUNT(*) OVER () AS cnt
			FROM employees e
			JOIN countries c ON e.country_id = c.id
			WHERE c.name = ? AND e.is_active = 1
		) WHERE rn = (cnt + 1) / 2`, country,
	).Scan(&sr.Median)
	if err == sql.ErrNoRows {
		sr.Median = 0
	} else if err != nil {
		return nil, fmt.Errorf("median salary: %w", err)
	}

	return sr, nil
}

func (r *employeeRepo) GetSalaryByTitle(ctx context.Context, country, jobTitle string) (*model.SalaryByTitle, error) {
	sbt := &model.SalaryByTitle{Country: country, JobTitle: jobTitle}

	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG(e.salary), 0), COALESCE(MIN(e.salary), 0),
		       COALESCE(MAX(e.salary), 0), COUNT(*)
		FROM employees e
		JOIN countries  c  ON e.country_id   = c.id
		JOIN job_titles jt ON e.job_title_id = jt.id
		WHERE c.name = ? AND jt.name = ? AND e.is_active = 1`, country, jobTitle,
	).Scan(&sbt.Average, &sbt.Min, &sbt.Max, &sbt.Count)
	if err != nil {
		return nil, fmt.Errorf("salary by title: %w", err)
	}

	return sbt, nil
}

func (r *employeeRepo) GetDepartmentStats(ctx context.Context, country string) ([]model.DepartmentStats, error) {
	query := `
		SELECT d.name, AVG(e.salary), MIN(e.salary), MAX(e.salary), COUNT(*)
		FROM employees e
		JOIN job_titles  jt ON e.job_title_id  = jt.id
		JOIN departments d  ON jt.department_id = d.id`

	conds := []string{"e.is_active = 1"}
	var args []interface{}
	if country != "" {
		query += ` JOIN countries c ON e.country_id = c.id`
		conds = append(conds, "c.name = ?")
		args = append(args, country)
	}
	query += " WHERE " + strings.Join(conds, " AND ") + " GROUP BY d.name ORDER BY AVG(e.salary) DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("department stats: %w", err)
	}
	defer rows.Close()

	var stats []model.DepartmentStats
	for rows.Next() {
		var ds model.DepartmentStats
		if err := rows.Scan(&ds.Department, &ds.AverageSalary, &ds.MinSalary, &ds.MaxSalary, &ds.EmployeeCount); err != nil {
			return nil, fmt.Errorf("scan department stats: %w", err)
		}
		stats = append(stats, ds)
	}
	return stats, rows.Err()
}

func (r *employeeRepo) GetOrgSummary(ctx context.Context) (*model.OrgSummary, error) {
	summary := &model.OrgSummary{}

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(AVG(e.salary), 0),
		       COUNT(DISTINCT e.country_id),
		       COUNT(DISTINCT jt.department_id)
		FROM employees e
		JOIN job_titles jt ON e.job_title_id = jt.id
		WHERE e.is_active = 1`,
	).Scan(&summary.TotalEmployees, &summary.AverageSalary,
		&summary.TotalCountries, &summary.TotalDepartments)
	if err != nil {
		return nil, fmt.Errorf("org summary: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT c.name, COUNT(*), AVG(e.salary)
		FROM employees e
		JOIN countries c ON e.country_id = c.id
		WHERE e.is_active = 1
		GROUP BY c.name
		ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("country breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ch model.CountryHeadcount
		if err := rows.Scan(&ch.Country, &ch.EmployeeCount, &ch.AverageSalary); err != nil {
			return nil, fmt.Errorf("scan country headcount: %w", err)
		}
		summary.CountryBreakdown = append(summary.CountryBreakdown, ch)
	}
	return summary, rows.Err()
}
