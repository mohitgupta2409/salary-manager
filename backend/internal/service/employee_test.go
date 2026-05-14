package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// mockEmployeeRepo implements repository.EmployeeRepository.
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
	if !ok {
		return errors.New("employee not found")
	}
	if !e.IsActive {
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

// mockCountryRepo + mockJobTitleRepo for FK validation
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
func (m *mockCountryRepo) GetByCode(_ context.Context, _ string) (*model.Country, error) {
	return nil, nil
}

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

func newSvc() *EmployeeService {
	return NewEmployeeService(newMockEmployeeRepo(), newMockCountryRepo(), newMockJobTitleRepo())
}

func validEmployee() *model.Employee {
	return &model.Employee{
		FirstName:  "Jane",
		LastName:   "Doe",
		Email:      "jane@example.com",
		JobTitleID: 1,
		CountryID:  1,
		Salary:     100000,
		Address:    "123 Main St",
		JoinDate:   time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestService_Create_Success(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	if err := svc.Create(context.Background(), emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if emp.ID == 0 {
		t.Error("Create() should assign ID")
	}
}

func TestService_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*model.Employee)
		wantErr string
	}{
		{"empty first name", func(e *model.Employee) { e.FirstName = "" }, "first name is required"},
		{"empty last name", func(e *model.Employee) { e.LastName = "" }, "last name is required"},
		{"empty email", func(e *model.Employee) { e.Email = "" }, "email is required"},
		{"invalid email", func(e *model.Employee) { e.Email = "notanemail" }, "email must be valid"},
		{"missing job title id", func(e *model.Employee) { e.JobTitleID = 0 }, "job title is required"},
		{"missing country id", func(e *model.Employee) { e.CountryID = 0 }, "country is required"},
		{"negative salary", func(e *model.Employee) { e.Salary = -1 }, "salary must be non-negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc()
			emp := validEmployee()
			tt.modify(emp)

			err := svc.Create(context.Background(), emp)
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestService_Create_InactiveCountry(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	emp.CountryID = 2 // inactive in mock

	err := svc.Create(context.Background(), emp)
	if err == nil || err.Error() != "country does not exist or is inactive" {
		t.Errorf("error = %v, want 'country does not exist or is inactive'", err)
	}
}

func TestService_Create_NonExistentCountry(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	emp.CountryID = 999

	err := svc.Create(context.Background(), emp)
	if err == nil || err.Error() != "country does not exist or is inactive" {
		t.Errorf("error = %v", err)
	}
}

func TestService_Create_InactiveJobTitle(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	emp.JobTitleID = 2 // inactive

	err := svc.Create(context.Background(), emp)
	if err == nil || err.Error() != "job title does not exist or is inactive" {
		t.Errorf("error = %v", err)
	}
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	svc := newSvc()

	if err := svc.Create(context.Background(), validEmployee()); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	dup := validEmployee()
	dup.FirstName = "Different"
	err := svc.Create(context.Background(), dup)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("error = %v, want ErrDuplicateEmail", err)
	}
}

func TestService_Create_NormalizesInput(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	emp.Email = "  JANE@Example.COM  "
	emp.FirstName = "  Jane  "
	emp.LastName = "  Doe  "
	emp.Address = "  123 Main St  "

	if err := svc.Create(context.Background(), emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if emp.Email != "jane@example.com" {
		t.Errorf("Email = %q, want lowercase trimmed", emp.Email)
	}
	if emp.FirstName != "Jane" || emp.LastName != "Doe" {
		t.Errorf("name not trimmed: %q %q", emp.FirstName, emp.LastName)
	}
	if emp.Address != "123 Main St" {
		t.Errorf("Address = %q", emp.Address)
	}
}

func TestService_GetByID_Success(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	_ = svc.Create(context.Background(), emp)

	got, err := svc.GetByID(context.Background(), emp.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.FirstName != "Jane" {
		t.Errorf("FirstName = %q", got.FirstName)
	}
}

func TestService_GetByID_InvalidID(t *testing.T) {
	svc := newSvc()
	for _, id := range []int64{0, -1} {
		_, err := svc.GetByID(context.Background(), id)
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("GetByID(%d) error = %v, want ErrInvalidInput", id, err)
		}
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestService_Update_Success(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	_ = svc.Create(context.Background(), emp)

	emp.Salary = 130000
	if err := svc.Update(context.Background(), emp); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestService_Update_InvalidID(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	emp.ID = 0
	if err := svc.Update(context.Background(), emp); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	emp.ID = 999
	if err := svc.Update(context.Background(), emp); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestService_Delete_SoftDelete(t *testing.T) {
	svc := newSvc()
	emp := validEmployee()
	_ = svc.Create(context.Background(), emp)

	if err := svc.Delete(context.Background(), emp.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	// Soft-deleted employee should still exist but not in list
	got, err := svc.GetByID(context.Background(), emp.ID)
	if err != nil {
		t.Fatalf("GetByID after Delete error = %v", err)
	}
	if got == nil || got.IsActive {
		t.Error("after soft delete, employee should still exist with IsActive=false")
	}
}

func TestService_Delete_InvalidID(t *testing.T) {
	svc := newSvc()
	if err := svc.Delete(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	svc := newSvc()
	if err := svc.Delete(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

func TestService_List_DefaultPagination(t *testing.T) {
	svc := newSvc()
	result, err := svc.List(context.Background(), model.EmployeeFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Page != 1 || result.Limit != 20 {
		t.Errorf("default pagination = %+v, want page=1 limit=20", result)
	}
}

func TestService_GetSalaryRange_EmptyCountry(t *testing.T) {
	svc := newSvc()
	if _, err := svc.GetSalaryRangeByCountry(context.Background(), ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestService_GetSalaryByTitle_EmptyInputs(t *testing.T) {
	svc := newSvc()
	if _, err := svc.GetSalaryByTitle(context.Background(), "", "Engineer"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v", err)
	}
	if _, err := svc.GetSalaryByTitle(context.Background(), "USA", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v", err)
	}
}

func TestEmployee_FullName(t *testing.T) {
	tests := []struct {
		first, last, want string
	}{
		{"Jane", "Doe", "Jane Doe"},
		{"", "Doe", "Doe"},
		{"Jane", "", "Jane"},
	}
	for _, tt := range tests {
		e := model.Employee{FirstName: tt.first, LastName: tt.last}
		if got := e.FullName(); got != tt.want {
			t.Errorf("FullName(%q,%q) = %q, want %q", tt.first, tt.last, got, tt.want)
		}
	}
}
