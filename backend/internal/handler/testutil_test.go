package handler

// This file contains the shared test infrastructure for the handler
// package: in-memory mock repositories, the canonical setupRouter that
// wires the full HTTP stack, and small fixture helpers used across the
// per-entity *_test.go files.
//
// Each per-entity test file (country_test.go, department_test.go,
// job_title_test.go, employee_test.go, insights_test.go) only contains
// its own scenarios; common setup lives here.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/service"
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
	// Mimic the real repo populating denormalised fields via JOINs.
	emp.Country = "USA"
	emp.Currency = "USD"
	emp.JobTitle = "Engineer"
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
	emp.Country = "USA"
	emp.Currency = "USD"
	emp.JobTitle = "Engineer"
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
	var out []model.Employee
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
	return &model.SalaryRange{Country: c, Min: 50000, Max: 150000, Average: 100000, Count: 5}, nil
}

func (m *mockEmployeeRepo) GetSalaryByTitle(_ context.Context, c, t string) (*model.SalaryByTitle, error) {
	return &model.SalaryByTitle{Country: c, JobTitle: t, Average: 110000, Count: 3}, nil
}

func (m *mockEmployeeRepo) GetDepartmentStats(_ context.Context, _ string) ([]model.DepartmentStats, error) {
	return []model.DepartmentStats{{Department: "Eng", AverageSalary: 120000, EmployeeCount: 10}}, nil
}

func (m *mockEmployeeRepo) GetOrgSummary(_ context.Context) (*model.OrgSummary, error) {
	return &model.OrgSummary{TotalEmployees: 50, AverageSalary: 95000}, nil
}

// =============================================================================
// mockCountryRepo
// =============================================================================

type mockCountryRepo struct {
	byID map[int64]*model.Country
}

func newMockCountryRepo() *mockCountryRepo {
	return &mockCountryRepo{byID: map[int64]*model.Country{
		1: {ID: 1, Name: "USA", Code: "US", Currency: "USD", IsActive: true},
	}}
}

func (m *mockCountryRepo) Create(_ context.Context, c *model.Country) error {
	c.ID = int64(len(m.byID) + 1)
	c.IsActive = true
	m.byID[c.ID] = c
	return nil
}

func (m *mockCountryRepo) List(_ context.Context, req model.CountryListRequest) (*model.CountryListResult, error) {
	out := []model.Country{}
	for _, c := range m.byID {
		out = append(out, *c)
	}
	return &model.CountryListResult{
		Countries: out,
		Total:     int64(len(out)),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
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
	return &mockDeptRepo{byID: map[int64]*model.Department{
		1: {ID: 1, Name: "Engineering", IsActive: true},
	}}
}

func (m *mockDeptRepo) Create(_ context.Context, d *model.Department) error {
	d.ID = int64(len(m.byID) + 1)
	d.IsActive = true
	m.byID[d.ID] = d
	return nil
}

func (m *mockDeptRepo) List(_ context.Context, req model.DepartmentListRequest) (*model.DepartmentListResult, error) {
	out := []model.Department{}
	for _, d := range m.byID {
		out = append(out, *d)
	}
	return &model.DepartmentListResult{
		Departments: out,
		Total:       int64(len(out)),
		Limit:       req.Limit,
		Offset:      req.Offset,
	}, nil
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
	return &mockJobTitleRepo{byID: map[int64]*model.JobTitle{
		1: {ID: 1, Name: "Engineer", DepartmentID: 1, Department: "Engineering", IsActive: true},
	}}
}

func (m *mockJobTitleRepo) Create(_ context.Context, jt *model.JobTitle) error {
	jt.ID = int64(len(m.byID) + 1)
	jt.IsActive = true
	m.byID[jt.ID] = jt
	return nil
}

func (m *mockJobTitleRepo) List(_ context.Context, req model.JobTitleListRequest) (*model.JobTitleListResult, error) {
	out := []model.JobTitle{}
	for _, jt := range m.byID {
		out = append(out, *jt)
	}
	return &model.JobTitleListResult{
		JobTitles: out,
		Total:     int64(len(out)),
		Limit:     req.Limit,
		Offset:    req.Offset,
	}, nil
}

func (m *mockJobTitleRepo) GetByID(_ context.Context, id int64) (*model.JobTitle, error) {
	jt, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return jt, nil
}

// =============================================================================
// Router & request helpers
// =============================================================================

// setupRouter wires up the complete HTTP API on top of the mock
// repositories. Tests should treat the returned http.Handler as a black
// box and exercise it via httptest.NewRequest + ServeHTTP.
func setupRouter() http.Handler {
	empRepo := newMockEmployeeRepo()
	countryRepo := newMockCountryRepo()
	deptRepo := newMockDeptRepo()
	jtRepo := newMockJobTitleRepo()

	empSvc := service.NewEmployeeService(empRepo, countryRepo, jtRepo)
	countrySvc := service.NewCountryService(countryRepo)
	deptSvc := service.NewDepartmentService(deptRepo)
	jtSvc := service.NewJobTitleService(jtRepo, deptRepo)

	r := chi.NewRouter()
	r.Mount("/api/employees", NewEmployeeHandler(empSvc).Routes())
	r.Mount("/api/countries", NewCountryHandler(countrySvc).Routes())
	r.Mount("/api/departments", NewDepartmentHandler(deptSvc).Routes())
	r.Mount("/api/job-titles", NewJobTitleHandler(jtSvc).Routes())
	r.Mount("/api/insights", InsightsRoutes(empSvc))
	return r
}

// validEmployeeJSON returns a body that matches dto.EmployeeCreateRequest
// and passes service-level validation against the default mock fixtures.
func validEmployeeJSON() []byte {
	body := map[string]interface{}{
		"first_name":   "Jane",
		"last_name":    "Doe",
		"email":        "jane@example.com",
		"job_title_id": 1,
		"country_id":   1,
		"salary":       100000,
		"address":      "123 Main St",
		"join_date":    time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	b, _ := json.Marshal(body)
	return b
}

// do performs an HTTP request against router and returns the recorder.
// All test files use this so the JSON Content-Type header and body
// plumbing are not repeated on every call site.
func do(router http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, r)
	return rec
}

// decodeJSON unmarshals a recorder's body into v, failing the test on
// malformed JSON.
func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode JSON failed: %v\nbody: %s", err, rec.Body.String())
	}
}
