package handler

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

// ----- mock repos -----

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
	c := *emp
	m.employees[emp.ID] = &c
	return nil
}

func (m *mockEmployeeRepo) SoftDelete(_ context.Context, id int64) error {
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
	return &model.EmployeeListResult{Employees: out, Total: int64(len(out)), Page: f.Page, Limit: f.Limit}, nil
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
func (m *mockCountryRepo) List(_ context.Context, _ bool) ([]model.Country, error) {
	out := []model.Country{}
	for _, c := range m.byID {
		out = append(out, *c)
	}
	return out, nil
}
func (m *mockCountryRepo) GetByID(_ context.Context, id int64) (*model.Country, error) {
	c, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}
func (m *mockCountryRepo) GetByCode(_ context.Context, _ string) (*model.Country, error) { return nil, nil }

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
func (m *mockDeptRepo) List(_ context.Context, _ bool) ([]model.Department, error) {
	out := []model.Department{}
	for _, d := range m.byID {
		out = append(out, *d)
	}
	return out, nil
}
func (m *mockDeptRepo) GetByID(_ context.Context, id int64) (*model.Department, error) {
	d, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return d, nil
}

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
func (m *mockJobTitleRepo) List(_ context.Context, _ int64, _ bool) ([]model.JobTitle, error) {
	out := []model.JobTitle{}
	for _, jt := range m.byID {
		out = append(out, *jt)
	}
	return out, nil
}
func (m *mockJobTitleRepo) GetByID(_ context.Context, id int64) (*model.JobTitle, error) {
	jt, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return jt, nil
}

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

func validEmployeeJSON() []byte {
	emp := map[string]interface{}{
		"first_name":   "Jane",
		"last_name":    "Doe",
		"email":        "jane@example.com",
		"job_title_id": 1,
		"country_id":   1,
		"salary":       100000,
		"address":      "123 Main St",
		"join_date":    time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	b, _ := json.Marshal(emp)
	return b
}

// ----- employee handler tests -----

func TestHandler_CreateEmployee(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var emp model.Employee
	json.NewDecoder(rec.Body).Decode(&emp)
	if emp.ID == 0 || emp.FirstName != "Jane" {
		t.Errorf("unexpected response: %+v", emp)
	}
}

func TestHandler_CreateEmployee_InvalidJSON(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateEmployee_ValidationError(t *testing.T) {
	router := setupRouter()
	body, _ := json.Marshal(map[string]interface{}{"first_name": ""})
	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_CreateEmployee_DuplicateEmail(t *testing.T) {
	router := setupRouter()

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if i == 1 && rec.Code != http.StatusConflict {
			t.Errorf("second create status = %d, want 409", rec.Code)
		}
	}
}

func TestHandler_GetByID(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	req = httptest.NewRequest("GET", "/api/employees/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/employees/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/employees/abc", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandler_UpdateEmployee(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	updated, _ := json.Marshal(map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith", "email": "jane@example.com",
		"job_title_id": 1, "country_id": 1, "salary": 130000,
		"join_date": time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	req = httptest.NewRequest("PUT", "/api/employees/1", bytes.NewReader(updated))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_DeleteEmployee_SoftDelete(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	req = httptest.NewRequest("DELETE", "/api/employees/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("delete status = %d", rec.Code)
	}

	// List should not include the soft-deleted employee
	req = httptest.NewRequest("GET", "/api/employees", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var result model.EmployeeListResult
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 0 {
		t.Errorf("after soft delete, list total = %d, want 0", result.Total)
	}
}

func TestHandler_ListEmployees(t *testing.T) {
	router := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	req = httptest.NewRequest("GET", "/api/employees?page=1&limit=10", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	var result model.EmployeeListResult
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}

// ----- insights handler tests -----

func TestHandler_InsightsSalaryRange(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/insights/salary-range?country=USA", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandler_InsightsSummary(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/insights/summary", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var summary model.OrgSummary
	json.NewDecoder(rec.Body).Decode(&summary)
	if summary.TotalEmployees != 50 {
		t.Errorf("TotalEmployees = %d, want 50", summary.TotalEmployees)
	}
}

// ----- reference handler tests -----

func TestHandler_ListCountries(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/countries", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	var countries []model.Country
	json.NewDecoder(rec.Body).Decode(&countries)
	if len(countries) != 1 {
		t.Errorf("got %d countries, want 1", len(countries))
	}
}

func TestHandler_CreateCountry(t *testing.T) {
	router := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": "Brazil", "code": "BR", "currency": "BRL"})
	req := httptest.NewRequest("POST", "/api/countries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d. Body: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_CreateCountry_InvalidName(t *testing.T) {
	router := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": "", "code": "BR", "currency": "BRL"})
	req := httptest.NewRequest("POST", "/api/countries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandler_ListDepartments(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/departments", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandler_ListJobTitles(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/job-titles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandler_ListJobTitles_FilterByDept(t *testing.T) {
	router := setupRouter()
	req := httptest.NewRequest("GET", "/api/job-titles?department_id=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
}
