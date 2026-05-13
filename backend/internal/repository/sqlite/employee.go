package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type employeeRepo struct {
	db *sql.DB
}

func NewEmployeeRepository(db *sql.DB) repository.EmployeeRepository {
	return &employeeRepo{db: db}
}

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return db, nil
}

func runMigrations(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		job_title TEXT NOT NULL,
		department TEXT NOT NULL,
		country TEXT NOT NULL,
		salary REAL NOT NULL CHECK(salary >= 0),
		currency TEXT NOT NULL DEFAULT 'USD',
		join_date DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_employees_country ON employees(country);
	CREATE INDEX IF NOT EXISTS idx_employees_job_title ON employees(job_title);
	CREATE INDEX IF NOT EXISTS idx_employees_department ON employees(department);
	CREATE INDEX IF NOT EXISTS idx_employees_email ON employees(email);
	`
	_, err := db.Exec(query)
	return err
}

func (r *employeeRepo) Create(ctx context.Context, emp *model.Employee) error {
	now := time.Now().UTC()
	emp.CreatedAt = now
	emp.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO employees (full_name, email, job_title, department, country, salary, currency, join_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emp.FullName, emp.Email, emp.JobTitle, emp.Department, emp.Country,
		emp.Salary, emp.Currency, emp.JoinDate, emp.CreatedAt, emp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert employee: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	emp.ID = id
	return nil
}

func (r *employeeRepo) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	emp := &model.Employee{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, full_name, email, job_title, department, country, salary, currency, join_date, created_at, updated_at
		FROM employees WHERE id = ?`, id,
	).Scan(&emp.ID, &emp.FullName, &emp.Email, &emp.JobTitle, &emp.Department,
		&emp.Country, &emp.Salary, &emp.Currency, &emp.JoinDate, &emp.CreatedAt, &emp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get employee by id: %w", err)
	}
	return emp, nil
}

func (r *employeeRepo) Update(ctx context.Context, emp *model.Employee) error {
	emp.UpdatedAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, `
		UPDATE employees
		SET full_name = ?, email = ?, job_title = ?, department = ?, country = ?,
		    salary = ?, currency = ?, join_date = ?, updated_at = ?
		WHERE id = ?`,
		emp.FullName, emp.Email, emp.JobTitle, emp.Department, emp.Country,
		emp.Salary, emp.Currency, emp.JoinDate, emp.UpdatedAt, emp.ID,
	)
	if err != nil {
		return fmt.Errorf("update employee: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("employee with id %d not found", emp.ID)
	}
	return nil
}

func (r *employeeRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM employees WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete employee: %w", err)
	}
	rows, err := result.RowsAffected()
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

	var conditions []string
	var args []interface{}

	if filter.Country != "" {
		conditions = append(conditions, "country = ?")
		args = append(args, filter.Country)
	}
	if filter.JobTitle != "" {
		conditions = append(conditions, "job_title = ?")
		args = append(args, filter.JobTitle)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(full_name LIKE ? OR email LIKE ?)")
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM employees " + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count employees: %w", err)
	}

	offset := (filter.Page - 1) * filter.Limit
	dataQuery := fmt.Sprintf(`
		SELECT id, full_name, email, job_title, department, country, salary, currency, join_date, created_at, updated_at
		FROM employees %s ORDER BY id DESC LIMIT ? OFFSET ?`, whereClause)

	dataArgs := append(args, filter.Limit, offset)
	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	var employees []model.Employee
	for rows.Next() {
		var emp model.Employee
		if err := rows.Scan(&emp.ID, &emp.FullName, &emp.Email, &emp.JobTitle, &emp.Department,
			&emp.Country, &emp.Salary, &emp.Currency, &emp.JoinDate, &emp.CreatedAt, &emp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		employees = append(employees, emp)
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
		SELECT COALESCE(MIN(salary), 0), COALESCE(MAX(salary), 0),
		       COALESCE(AVG(salary), 0), COUNT(*)
		FROM employees WHERE country = ?`, country,
	).Scan(&sr.Min, &sr.Max, &sr.Average, &sr.Count)
	if err != nil {
		return nil, fmt.Errorf("salary range by country: %w", err)
	}

	// Calculate median using window function
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(salary, 0) FROM (
			SELECT salary, ROW_NUMBER() OVER (ORDER BY salary) as rn,
			       COUNT(*) OVER () as cnt
			FROM employees WHERE country = ?
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
		SELECT COALESCE(AVG(salary), 0), COALESCE(MIN(salary), 0),
		       COALESCE(MAX(salary), 0), COUNT(*)
		FROM employees WHERE country = ? AND job_title = ?`, country, jobTitle,
	).Scan(&sbt.Average, &sbt.Min, &sbt.Max, &sbt.Count)
	if err != nil {
		return nil, fmt.Errorf("salary by title: %w", err)
	}

	return sbt, nil
}

func (r *employeeRepo) GetDepartmentStats(ctx context.Context, country string) ([]model.DepartmentStats, error) {
	query := `
		SELECT department, AVG(salary), MIN(salary), MAX(salary), COUNT(*)
		FROM employees`

	var args []interface{}
	if country != "" {
		query += " WHERE country = ?"
		args = append(args, country)
	}
	query += " GROUP BY department ORDER BY AVG(salary) DESC"

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
		SELECT COUNT(*), COALESCE(AVG(salary), 0),
		       COUNT(DISTINCT country), COUNT(DISTINCT department)
		FROM employees`,
	).Scan(&summary.TotalEmployees, &summary.AverageSalary,
		&summary.TotalCountries, &summary.TotalDepartments)
	if err != nil {
		return nil, fmt.Errorf("org summary: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT country, COUNT(*), AVG(salary)
		FROM employees GROUP BY country ORDER BY COUNT(*) DESC`)
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
