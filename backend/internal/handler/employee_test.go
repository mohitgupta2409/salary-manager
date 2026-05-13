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

// mockRepo for handler tests
type mockRepo struct {
	employees map[int64]*model.Employee
	nextID    int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{employees: make(map[int64]*model.Employee), nextID: 1}
}

func (m *mockRepo) Create(_ context.Context, emp *model.Employee) error {
	for _, e := range m.employees {
		if e.Email == emp.Email {
			return errors.New("UNIQUE constraint failed: employees.email")
		}
	}
	emp.ID = m.nextID
	m.nextID++
	stored := *emp
	m.employees[emp.ID] = &stored
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id int64) (*model.Employee, error) {
	e, ok := m.employees[id]
	if !ok {
		return nil, nil
	}
	copy := *e
	return &copy, nil
}

func (m *mockRepo) Update(_ context.Context, emp *model.Employee) error {
	if _, ok := m.employees[emp.ID]; !ok {
		return errors.New("employee not found")
	}
	stored := *emp
	m.employees[emp.ID] = &stored
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id int64) error {
	if _, ok := m.employees[id]; !ok {
		return errors.New("employee not found")
	}
	delete(m.employees, id)
	return nil
}

func (m *mockRepo) List(_ context.Context, f model.EmployeeFilter) (*model.EmployeeListResult, error) {
	var all []model.Employee
	for _, e := range m.employees {
		all = append(all, *e)
	}
	return &model.EmployeeListResult{Employees: all, Total: int64(len(all)), Page: f.Page, Limit: f.Limit}, nil
}

func (m *mockRepo) GetSalaryRangeByCountry(_ context.Context, c string) (*model.SalaryRange, error) {
	return &model.SalaryRange{Country: c, Min: 50000, Max: 150000, Average: 100000, Count: 5}, nil
}

func (m *mockRepo) GetSalaryByTitle(_ context.Context, c, t string) (*model.SalaryByTitle, error) {
	return &model.SalaryByTitle{Country: c, JobTitle: t, Average: 110000, Count: 3}, nil
}

func (m *mockRepo) GetDepartmentStats(_ context.Context, _ string) ([]model.DepartmentStats, error) {
	return []model.DepartmentStats{{Department: "Eng", AverageSalary: 120000, EmployeeCount: 10}}, nil
}

func (m *mockRepo) GetOrgSummary(_ context.Context) (*model.OrgSummary, error) {
	return &model.OrgSummary{TotalEmployees: 50, AverageSalary: 95000}, nil
}

func setupHandler() (*EmployeeHandler, *service.EmployeeService) {
	repo := newMockRepo()
	svc := service.NewEmployeeService(repo)
	h := NewEmployeeHandler(svc)
	return h, svc
}

func setupRouter() (http.Handler, *service.EmployeeService) {
	h, svc := setupHandler()
	r := chi.NewRouter()
	r.Mount("/api/employees", h.Routes())
	r.Mount("/api/insights", InsightsRoutes(svc))
	return r, svc
}

func validEmployeeJSON() []byte {
	emp := map[string]interface{}{
		"full_name":  "Jane Doe",
		"email":      "jane@example.com",
		"job_title":  "Engineer",
		"department": "Engineering",
		"country":    "USA",
		"salary":     100000,
		"currency":   "USD",
		"join_date":  time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	b, _ := json.Marshal(emp)
	return b
}

func TestHandler_CreateEmployee(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d. Body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var emp model.Employee
	if err := json.NewDecoder(rec.Body).Decode(&emp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if emp.ID == 0 {
		t.Error("response should have non-zero ID")
	}
	if emp.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want Jane Doe", emp.FullName)
	}
}

func TestHandler_CreateEmployee_InvalidJSON(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_CreateEmployee_ValidationError(t *testing.T) {
	router, _ := setupRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"full_name": "",
		"email":     "jane@example.com",
	})
	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetByID(t *testing.T) {
	router, _ := setupRouter()

	// Create first
	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Get by ID
	req = httptest.NewRequest("GET", "/api/employees/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("GET", "/api/employees/999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("GET", "/api/employees/abc", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandler_UpdateEmployee(t *testing.T) {
	router, _ := setupRouter()

	// Create
	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Update
	updated, _ := json.Marshal(map[string]interface{}{
		"full_name":  "Jane Smith",
		"email":      "jane@example.com",
		"job_title":  "Senior Engineer",
		"department": "Engineering",
		"country":    "USA",
		"salary":     130000,
		"currency":   "USD",
		"join_date":  time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	req = httptest.NewRequest("PUT", "/api/employees/1", bytes.NewReader(updated))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d. Body: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var emp model.Employee
	json.NewDecoder(rec.Body).Decode(&emp)
	if emp.Salary != 130000 {
		t.Errorf("Salary = %f, want 130000", emp.Salary)
	}
}

func TestHandler_DeleteEmployee(t *testing.T) {
	router, _ := setupRouter()

	// Create
	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// Delete
	req = httptest.NewRequest("DELETE", "/api/employees/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Verify gone
	req = httptest.NewRequest("GET", "/api/employees/1", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("after delete, status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandler_ListEmployees(t *testing.T) {
	router, _ := setupRouter()

	// Create
	req := httptest.NewRequest("POST", "/api/employees", bytes.NewReader(validEmployeeJSON()))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// List
	req = httptest.NewRequest("GET", "/api/employees?page=1&limit=10", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var result model.EmployeeListResult
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
}

func TestHandler_InsightsSalaryRange(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("GET", "/api/insights/salary-range?country=USA", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var sr model.SalaryRange
	json.NewDecoder(rec.Body).Decode(&sr)
	if sr.Country != "USA" {
		t.Errorf("Country = %q, want USA", sr.Country)
	}
}

func TestHandler_InsightsSalaryByTitle(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("GET", "/api/insights/salary-by-title?country=USA&job_title=Engineer", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHandler_InsightsSummary(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("GET", "/api/insights/summary", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var summary model.OrgSummary
	json.NewDecoder(rec.Body).Decode(&summary)
	if summary.TotalEmployees != 50 {
		t.Errorf("TotalEmployees = %d, want 50", summary.TotalEmployees)
	}
}

func TestHandler_InsightsDepartmentStats(t *testing.T) {
	router, _ := setupRouter()

	req := httptest.NewRequest("GET", "/api/insights/department-stats?country=USA", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
