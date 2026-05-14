// Package repository defines data-access interfaces for all domain entities.
// Implementations live in subpackages (e.g., sqlite).
package repository

import (
	"context"

	"github.com/salary-manager/backend/internal/model"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) error
	GetByID(ctx context.Context, id int64) (*model.Employee, error)
	Update(ctx context.Context, emp *model.Employee) error
	SoftDelete(ctx context.Context, id int64) error
	List(ctx context.Context, filter model.EmployeeFilter) (*model.EmployeeListResult, error)

	// Salary insights (operate on active employees only)
	GetSalaryRangeByCountry(ctx context.Context, country string) (*model.SalaryRange, error)
	GetSalaryByTitle(ctx context.Context, country, jobTitle string) (*model.SalaryByTitle, error)
	GetDepartmentStats(ctx context.Context, country string) ([]model.DepartmentStats, error)
	GetOrgSummary(ctx context.Context) (*model.OrgSummary, error)
}

type CountryRepository interface {
	Create(ctx context.Context, c *model.Country) error
	List(ctx context.Context, includeInactive bool) ([]model.Country, error)
	GetByID(ctx context.Context, id int64) (*model.Country, error)
	GetByCode(ctx context.Context, code string) (*model.Country, error)
}

type DepartmentRepository interface {
	Create(ctx context.Context, d *model.Department) error
	List(ctx context.Context, includeInactive bool) ([]model.Department, error)
	GetByID(ctx context.Context, id int64) (*model.Department, error)
}

type JobTitleRepository interface {
	Create(ctx context.Context, jt *model.JobTitle) error
	List(ctx context.Context, departmentID int64, includeInactive bool) ([]model.JobTitle, error)
	GetByID(ctx context.Context, id int64) (*model.JobTitle, error)
}
