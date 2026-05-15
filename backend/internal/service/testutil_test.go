package service

// This file holds the shared test infrastructure for the service package:
// in-memory mock repositories that satisfy the repository.* interfaces, and
// small constructors used by the per-entity *_test.go files.
//
// Each per-entity test file focuses on the scenarios for its service; it
// reaches into this file for the mocks rather than redefining them.

import (
	"context"
	"errors"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
)

// =============================================================================
// mockEmployeeRepo
// =============================================================================

type mockEmployeeRepo struct {
	employees map[int64]*model.Employee
	nextID    int64
}

func newMockEmployeeRepo() *mockEmployeeRepo {
	return &mockEmployeeRepo{employees: make(map[int64]*model.Employee), nextID: 1}
}

func (m *mockEmployeeRepo) Create(_ context.Context, emp *model.Employee) error {
	for _, e := range m.employees {
		if e.Email == emp.Email {
			return errors.New("UNIQUE constraint failed: employees.email")
		}
	}
	emp.ID = m.nextID
	m.nextID++
	emp.IsActive = true
	// Populate denormalised fields the way the real repo would, so DTO
	// conversion in the service layer reflects realistic data.
	emp.Country = "United States"
	emp.Currency = "USD"
	emp.JobTitle = "Software Engineer"
	emp.Department = "Engineering"
	stored := *emp
	m.employees[emp.ID] = &stored
	return nil
}

func (m *mockEmployeeRepo) GetByID(_ context.Context, id int64) (*model.Employee, error) {
	e, ok := m.employees[id]
	if !ok {
		return nil, nil
	}
	c := *e
	return &c, nil
}

func (m *mockEmployeeRepo) Update(_ context.Context, emp *model.Employee) error {
	if _, ok := m.employees[emp.ID]; !ok {
		return errors.New("employee not found")
	}
	emp.Country = "United States"
	emp.Currency = "USD"
	emp.JobTitle = "Software Engineer"
	emp.Department = "Engineering"
	c := *emp
	m.employees[emp.ID] = &c
	return nil
}

func (m *mockEmployeeRepo) Delete(_ context.Context, id int64) error {
	e, ok := m.employees[id]
	if !ok || !e.IsActive {
		return errors.New("employee not found")
	}
	e.IsActive = false
	return nil
}

func (m *mockEmployeeRepo) List(_ context.Context, f model.EmployeeFilter) (*model.EmployeeListResult, error) {
	out := []model.Employee{}
	for _, e := range m.employees {
		if e.IsActive {
			out = append(out, *e)
		}
	}
	return &model.EmployeeListResult{
		Employees: out,
		Total:     int64(len(out)),
		Limit:     f.Limit,
		Offset:    f.Offset,
	}, nil
}

func (m *mockEmployeeRepo) GetSalaryRangeByCountry(_ context.Context, c string) (*model.SalaryRange, error) {
	return &model.SalaryRange{Country: c, Min: 50000, Max: 150000, Average: 100000, Count: 10}, nil
}

func (m *mockEmployeeRepo) GetSalaryByTitle(_ context.Context, c, t string) (*model.SalaryByTitle, error) {
	return &model.SalaryByTitle{Country: c, JobTitle: t, Average: 110000, Count: 5}, nil
}

func (m *mockEmployeeRepo) GetDepartmentStats(_ context.Context, _ string) ([]model.DepartmentStats, error) {
	return []model.DepartmentStats{{Department: "Engineering", AverageSalary: 120000, EmployeeCount: 50}}, nil
}

func (m *mockEmployeeRepo) GetOrgSummary(_ context.Context) (*model.OrgSummary, error) {
	return &model.OrgSummary{TotalEmployees: 100, AverageSalary: 95000}, nil
}

// =============================================================================
// mockCountryRepo
// =============================================================================

type mockCountryRepo struct {
	byID map[int64]*model.Country
}

func newMockCountryRepo() *mockCountryRepo {
	return &mockCountryRepo{
		byID: map[int64]*model.Country{
			1: {ID: 1, Name: "United States", Code: "US", Currency: "USD", IsActive: true},
			2: {ID: 2, Name: "Inactive Land", Code: "IL", Currency: "ILX", IsActive: false},
		},
	}
}

func (m *mockCountryRepo) Create(_ context.Context, c *model.Country) error {
	c.ID = int64(len(m.byID) + 1)
	c.IsActive = true
	m.byID[c.ID] = c
	return nil
}

func (m *mockCountryRepo) List(_ context.Context, _ model.CountryListRequest) (*model.CountryListResult, error) {
	out := []model.Country{}
	for _, c := range m.byID {
		out = append(out, *c)
	}
	return &model.CountryListResult{Countries: out, Total: int64(len(out))}, nil
}

func (m *mockCountryRepo) GetByID(_ context.Context, id int64) (*model.Country, error) {
	c, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockCountryRepo) GetByCode(_ context.Context, _ string) (*model.Country, error) {
	return nil, nil
}

// =============================================================================
// mockDeptRepo
// =============================================================================

type mockDeptRepo struct {
	byID map[int64]*model.Department
}

func newMockDeptRepo() *mockDeptRepo {
	return &mockDeptRepo{
		byID: map[int64]*model.Department{
			1: {ID: 1, Name: "Engineering", IsActive: true},
			2: {ID: 2, Name: "Inactive Dept", IsActive: false},
		},
	}
}

func (m *mockDeptRepo) Create(_ context.Context, d *model.Department) error {
	d.ID = int64(len(m.byID) + 1)
	d.IsActive = true
	m.byID[d.ID] = d
	return nil
}

func (m *mockDeptRepo) List(_ context.Context, _ model.DepartmentListRequest) (*model.DepartmentListResult, error) {
	out := []model.Department{}
	for _, d := range m.byID {
		out = append(out, *d)
	}
	return &model.DepartmentListResult{Departments: out, Total: int64(len(out))}, nil
}

func (m *mockDeptRepo) GetByID(_ context.Context, id int64) (*model.Department, error) {
	d, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return d, nil
}

// =============================================================================
// mockJobTitleRepo
// =============================================================================

type mockJobTitleRepo struct {
	byID map[int64]*model.JobTitle
}

func newMockJobTitleRepo() *mockJobTitleRepo {
	return &mockJobTitleRepo{
		byID: map[int64]*model.JobTitle{
			1: {ID: 1, Name: "Software Engineer", DepartmentID: 1, Department: "Engineering", IsActive: true},
			2: {ID: 2, Name: "Inactive Title", DepartmentID: 1, IsActive: false},
		},
	}
}

func (m *mockJobTitleRepo) Create(_ context.Context, jt *model.JobTitle) error {
	jt.ID = int64(len(m.byID) + 1)
	jt.IsActive = true
	m.byID[jt.ID] = jt
	return nil
}

func (m *mockJobTitleRepo) List(_ context.Context, _ model.JobTitleListRequest) (*model.JobTitleListResult, error) {
	out := []model.JobTitle{}
	for _, jt := range m.byID {
		out = append(out, *jt)
	}
	return &model.JobTitleListResult{JobTitles: out, Total: int64(len(out))}, nil
}

func (m *mockJobTitleRepo) GetByID(_ context.Context, id int64) (*model.JobTitle, error) {
	jt, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return jt, nil
}

// =============================================================================
// Helpers used by the per-entity test files
// =============================================================================

// newSvc wires up an EmployeeService with all-mock repositories.
func newSvc() *EmployeeService {
	return NewEmployeeService(newMockEmployeeRepo(), newMockCountryRepo(), newMockJobTitleRepo())
}

// validEmployeeRequest returns a fully-populated EmployeeCreateRequest that
// passes validation against the default mock fixtures (CountryID=1,
// JobTitleID=1 are the active rows in the mocks).
func validEmployeeRequest() *dto.EmployeeCreateRequest {
	return &dto.EmployeeCreateRequest{
		FirstName:  "Jane",
		LastName:   "Doe",
		Email:      "jane@example.com",
		JobTitleID: 1,
		CountryID:  1,
		Salary:     100000,
		Address:    "123 Main St",
	}
}
