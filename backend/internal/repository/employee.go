package repository

import (
	"context"

	"github.com/salary-manager/backend/internal/model"
)

type EmployeeRepository interface {
	Create(ctx context.Context, emp *model.Employee) error
	GetByID(ctx context.Context, id int64) (*model.Employee, error)
	Update(ctx context.Context, emp *model.Employee) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter model.EmployeeFilter) (*model.EmployeeListResult, error)

	// Salary insights
	GetSalaryRangeByCountry(ctx context.Context, country string) (*model.SalaryRange, error)
	GetSalaryByTitle(ctx context.Context, country, jobTitle string) (*model.SalaryByTitle, error)
	GetDepartmentStats(ctx context.Context, country string) ([]model.DepartmentStats, error)
	GetOrgSummary(ctx context.Context) (*model.OrgSummary, error)
}
