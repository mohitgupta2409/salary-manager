package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// mockRepo implements repository.EmployeeRepository for testing
type mockRepo struct {
	employees map[int64]*model.Employee
	nextID    int64
	createErr error
	updateErr error
	deleteErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		employees: make(map[int64]*model.Employee),
		nextID:    1,
	}
}

func (m *mockRepo) Create(_ context.Context, emp *model.Employee) error {
	if m.createErr != nil {
		return m.createErr
	}
	for _, existing := range m.employees {
		if existing.Email == emp.Email {
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
	emp, ok := m.employees[id]
	if !ok {
		return nil, nil
	}
	copy := *emp
	return &copy, nil
}

func (m *mockRepo) Update(_ context.Context, emp *model.Employee) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.employees[emp.ID]; !ok {
		return errors.New("employee not found")
	}
	stored := *emp
	m.employees[emp.ID] = &stored
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.employees[id]; !ok {
		return errors.New("employee not found")
	}
	delete(m.employees, id)
	return nil
}

func (m *mockRepo) List(_ context.Context, filter model.EmployeeFilter) (*model.EmployeeListResult, error) {
	var all []model.Employee
	for _, emp := range m.employees {
		all = append(all, *emp)
	}
	return &model.EmployeeListResult{
		Employees: all,
		Total:     int64(len(all)),
		Page:      filter.Page,
		Limit:     filter.Limit,
	}, nil
}

func (m *mockRepo) GetSalaryRangeByCountry(_ context.Context, country string) (*model.SalaryRange, error) {
	return &model.SalaryRange{Country: country, Min: 50000, Max: 150000, Average: 100000, Count: 10}, nil
}

func (m *mockRepo) GetSalaryByTitle(_ context.Context, country, title string) (*model.SalaryByTitle, error) {
	return &model.SalaryByTitle{Country: country, JobTitle: title, Average: 110000, Count: 5}, nil
}

func (m *mockRepo) GetDepartmentStats(_ context.Context, _ string) ([]model.DepartmentStats, error) {
	return []model.DepartmentStats{{Department: "Engineering", AverageSalary: 120000, EmployeeCount: 50}}, nil
}

func (m *mockRepo) GetOrgSummary(_ context.Context) (*model.OrgSummary, error) {
	return &model.OrgSummary{TotalEmployees: 100, AverageSalary: 95000}, nil
}

func validEmployee() *model.Employee {
	return &model.Employee{
		FullName:   "Jane Doe",
		Email:      "jane@example.com",
		JobTitle:   "Engineer",
		Department: "Engineering",
		Country:    "USA",
		Salary:     100000,
		Currency:   "USD",
		JoinDate:   time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestService_Create_Success(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())
	emp := validEmployee()

	err := svc.Create(context.Background(), emp)
	if err != nil {
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
		{"empty name", func(e *model.Employee) { e.FullName = "" }, "full name is required"},
		{"empty email", func(e *model.Employee) { e.Email = "" }, "email is required"},
		{"invalid email", func(e *model.Employee) { e.Email = "notanemail" }, "email must be valid"},
		{"empty job title", func(e *model.Employee) { e.JobTitle = "" }, "job title is required"},
		{"empty department", func(e *model.Employee) { e.Department = "" }, "department is required"},
		{"empty country", func(e *model.Employee) { e.Country = "" }, "country is required"},
		{"negative salary", func(e *model.Employee) { e.Salary = -1 }, "salary must be non-negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewEmployeeService(newMockRepo())
			emp := validEmployee()
			tt.modify(emp)

			err := svc.Create(context.Background(), emp)
			if err == nil {
				t.Fatal("Create() should return error")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Create() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	emp1 := validEmployee()
	if err := svc.Create(context.Background(), emp1); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	emp2 := validEmployee()
	emp2.FullName = "Another Jane"
	err := svc.Create(context.Background(), emp2)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("Create() error = %v, want ErrDuplicateEmail", err)
	}
}

func TestService_Create_NormalizesInput(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())
	emp := validEmployee()
	emp.Email = "  JANE@Example.COM  "
	emp.FullName = "  Jane Doe  "

	if err := svc.Create(context.Background(), emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if emp.Email != "jane@example.com" {
		t.Errorf("Email = %q, want lowercase trimmed", emp.Email)
	}
	if emp.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want trimmed", emp.FullName)
	}
}

func TestService_Create_DefaultCurrency(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())
	emp := validEmployee()
	emp.Currency = ""

	if err := svc.Create(context.Background(), emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if emp.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", emp.Currency)
	}
}

func TestService_GetByID_Success(t *testing.T) {
	repo := newMockRepo()
	svc := NewEmployeeService(repo)

	emp := validEmployee()
	_ = svc.Create(context.Background(), emp)

	got, err := svc.GetByID(context.Background(), emp.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want Jane Doe", got.FullName)
	}
}

func TestService_GetByID_InvalidID(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	_, err := svc.GetByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("GetByID(0) error = %v, want ErrInvalidInput", err)
	}

	_, err = svc.GetByID(context.Background(), -1)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("GetByID(-1) error = %v, want ErrInvalidInput", err)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetByID(999) error = %v, want ErrNotFound", err)
	}
}

func TestService_Update_Success(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	emp := validEmployee()
	_ = svc.Create(context.Background(), emp)

	emp.Salary = 130000
	emp.JobTitle = "Senior Engineer"
	err := svc.Update(context.Background(), emp)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
}

func TestService_Update_InvalidID(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())
	emp := validEmployee()
	emp.ID = 0

	err := svc.Update(context.Background(), emp)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Update() error = %v, want ErrInvalidInput", err)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())
	emp := validEmployee()
	emp.ID = 999

	err := svc.Update(context.Background(), emp)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestService_Delete_Success(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	emp := validEmployee()
	_ = svc.Create(context.Background(), emp)

	err := svc.Delete(context.Background(), emp.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = svc.GetByID(context.Background(), emp.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Error("GetByID after Delete should return ErrNotFound")
	}
}

func TestService_Delete_InvalidID(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	err := svc.Delete(context.Background(), 0)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("Delete(0) error = %v, want ErrInvalidInput", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	err := svc.Delete(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete(999) error = %v, want ErrNotFound", err)
	}
}

func TestService_List_DefaultPagination(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	result, err := svc.List(context.Background(), model.EmployeeFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Page != 1 {
		t.Errorf("Page = %d, want 1", result.Page)
	}
	if result.Limit != 20 {
		t.Errorf("Limit = %d, want 20", result.Limit)
	}
}

func TestService_GetSalaryRange_EmptyCountry(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	_, err := svc.GetSalaryRangeByCountry(context.Background(), "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("GetSalaryRangeByCountry('') error = %v, want ErrInvalidInput", err)
	}
}

func TestService_GetSalaryByTitle_EmptyInputs(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	_, err := svc.GetSalaryByTitle(context.Background(), "", "Engineer")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}

	_, err = svc.GetSalaryByTitle(context.Background(), "USA", "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestService_GetOrgSummary(t *testing.T) {
	svc := NewEmployeeService(newMockRepo())

	summary, err := svc.GetOrgSummary(context.Background())
	if err != nil {
		t.Fatalf("GetOrgSummary() error = %v", err)
	}
	if summary.TotalEmployees != 100 {
		t.Errorf("TotalEmployees = %d, want 100", summary.TotalEmployees)
	}
}
