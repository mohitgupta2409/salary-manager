package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

func setupTestDB(t *testing.T) *employeeRepo {
	t.Helper()
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &employeeRepo{db: db}
}

func newTestEmployee(name, country, title string, salary float64) *model.Employee {
	return &model.Employee{
		FullName:   name,
		Email:      name + "@test.com",
		JobTitle:   title,
		Department: "Engineering",
		Country:    country,
		Salary:     salary,
		Currency:   "USD",
		JoinDate:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreate(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	emp := newTestEmployee("Alice Johnson", "USA", "Software Engineer", 120000)
	err := repo.Create(ctx, emp)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if emp.ID == 0 {
		t.Error("Create() should set ID")
	}
	if emp.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	emp1 := newTestEmployee("Alice", "USA", "Engineer", 100000)
	emp1.Email = "alice@test.com"
	if err := repo.Create(ctx, emp1); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	emp2 := newTestEmployee("Alice Clone", "UK", "Manager", 120000)
	emp2.Email = "alice@test.com"
	err := repo.Create(ctx, emp2)
	if err == nil {
		t.Error("Create() with duplicate email should return error")
	}
}

func TestGetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	emp := newTestEmployee("Bob Smith", "India", "Product Manager", 90000)
	if err := repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(ctx, emp.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.FullName != "Bob Smith" {
		t.Errorf("GetByID().FullName = %q, want %q", got.FullName, "Bob Smith")
	}
	if got.Country != "India" {
		t.Errorf("GetByID().Country = %q, want %q", got.Country, "India")
	}
	if got.Salary != 90000 {
		t.Errorf("GetByID().Salary = %f, want %f", got.Salary, 90000.0)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	got, err := repo.GetByID(ctx, 9999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Error("GetByID() for non-existent ID should return nil")
	}
}

func TestUpdate(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	emp := newTestEmployee("Carol White", "Germany", "Designer", 85000)
	if err := repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	emp.Salary = 95000
	emp.JobTitle = "Senior Designer"
	if err := repo.Update(ctx, emp); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := repo.GetByID(ctx, emp.ID)
	if got.Salary != 95000 {
		t.Errorf("after Update(), Salary = %f, want %f", got.Salary, 95000.0)
	}
	if got.JobTitle != "Senior Designer" {
		t.Errorf("after Update(), JobTitle = %q, want %q", got.JobTitle, "Senior Designer")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	emp := &model.Employee{ID: 9999, FullName: "Nobody", Email: "no@one.com",
		JobTitle: "Ghost", Department: "None", Country: "Nowhere",
		Salary: 0, Currency: "USD", JoinDate: time.Now()}
	err := repo.Update(ctx, emp)
	if err == nil {
		t.Error("Update() for non-existent ID should return error")
	}
}

func TestDelete(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	emp := newTestEmployee("Dave Brown", "UK", "Analyst", 70000)
	if err := repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, emp.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, _ := repo.GetByID(ctx, emp.ID)
	if got != nil {
		t.Error("GetByID() after Delete() should return nil")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.Delete(ctx, 9999)
	if err == nil {
		t.Error("Delete() for non-existent ID should return error")
	}
}

func TestList_Pagination(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		emp := newTestEmployee(
			fmt.Sprintf("Employee_%d", i),
			"USA",
			"Engineer",
			float64(50000+i*1000),
		)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := repo.List(ctx, model.EmployeeFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 25 {
		t.Errorf("List().Total = %d, want 25", result.Total)
	}
	if len(result.Employees) != 10 {
		t.Errorf("List() returned %d employees, want 10", len(result.Employees))
	}

	// Page 3 should have 5 employees
	result, err = repo.List(ctx, model.EmployeeFilter{Page: 3, Limit: 10})
	if err != nil {
		t.Fatalf("List() page 3 error = %v", err)
	}
	if len(result.Employees) != 5 {
		t.Errorf("List() page 3 returned %d employees, want 5", len(result.Employees))
	}
}

func TestList_FilterByCountry(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	countries := []string{"USA", "USA", "India", "UK"}
	for i, c := range countries {
		emp := newTestEmployee(fmt.Sprintf("Emp_%d", i), c, "Engineer", 80000)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := repo.List(ctx, model.EmployeeFilter{Country: "USA", Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 2 {
		t.Errorf("List(country=USA).Total = %d, want 2", result.Total)
	}
}

func TestList_FilterBySearch(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	names := []string{"Alice Johnson", "Bob Smith", "Alice Walker"}
	for i, n := range names {
		emp := newTestEmployee(n, "USA", "Engineer", 80000)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := repo.List(ctx, model.EmployeeFilter{Search: "Alice", Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 2 {
		t.Errorf("List(search=Alice).Total = %d, want 2", result.Total)
	}
}

func TestGetSalaryRangeByCountry(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	salaries := []float64{60000, 80000, 100000, 120000, 140000}
	for i, s := range salaries {
		emp := newTestEmployee(fmt.Sprintf("Emp_%d", i), "India", "Engineer", s)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	sr, err := repo.GetSalaryRangeByCountry(ctx, "India")
	if err != nil {
		t.Fatalf("GetSalaryRangeByCountry() error = %v", err)
	}
	if sr.Min != 60000 {
		t.Errorf("Min = %f, want 60000", sr.Min)
	}
	if sr.Max != 140000 {
		t.Errorf("Max = %f, want 140000", sr.Max)
	}
	if sr.Average != 100000 {
		t.Errorf("Average = %f, want 100000", sr.Average)
	}
	if sr.Median != 100000 {
		t.Errorf("Median = %f, want 100000", sr.Median)
	}
	if sr.Count != 5 {
		t.Errorf("Count = %d, want 5", sr.Count)
	}
}

func TestGetSalaryByTitle(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	employees := []struct {
		title   string
		salary  float64
		country string
	}{
		{"Engineer", 100000, "USA"},
		{"Engineer", 120000, "USA"},
		{"Manager", 150000, "USA"},
		{"Engineer", 80000, "India"},
	}
	for i, e := range employees {
		emp := newTestEmployee(fmt.Sprintf("Emp_%d", i), e.country, e.title, e.salary)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	sbt, err := repo.GetSalaryByTitle(ctx, "USA", "Engineer")
	if err != nil {
		t.Fatalf("GetSalaryByTitle() error = %v", err)
	}
	if sbt.Average != 110000 {
		t.Errorf("Average = %f, want 110000", sbt.Average)
	}
	if sbt.Count != 2 {
		t.Errorf("Count = %d, want 2", sbt.Count)
	}
}

func TestGetDepartmentStats(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	data := []struct {
		dept    string
		salary  float64
		country string
	}{
		{"Engineering", 100000, "USA"},
		{"Engineering", 120000, "USA"},
		{"Marketing", 80000, "USA"},
		{"Engineering", 90000, "India"},
	}
	for i, d := range data {
		emp := newTestEmployee(fmt.Sprintf("Emp_%d", i), d.country, "Engineer", d.salary)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		emp.Department = d.dept
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	stats, err := repo.GetDepartmentStats(ctx, "USA")
	if err != nil {
		t.Fatalf("GetDepartmentStats() error = %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("got %d departments, want 2", len(stats))
	}
	// Engineering should be first (higher avg salary)
	if stats[0].Department != "Engineering" {
		t.Errorf("first department = %q, want Engineering", stats[0].Department)
	}
	if stats[0].EmployeeCount != 2 {
		t.Errorf("Engineering count = %d, want 2", stats[0].EmployeeCount)
	}
}

func TestGetOrgSummary(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	countries := []string{"USA", "USA", "India", "UK"}
	for i, c := range countries {
		emp := newTestEmployee(fmt.Sprintf("Emp_%d", i), c, "Engineer", 100000)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	summary, err := repo.GetOrgSummary(ctx)
	if err != nil {
		t.Fatalf("GetOrgSummary() error = %v", err)
	}
	if summary.TotalEmployees != 4 {
		t.Errorf("TotalEmployees = %d, want 4", summary.TotalEmployees)
	}
	if summary.TotalCountries != 3 {
		t.Errorf("TotalCountries = %d, want 3", summary.TotalCountries)
	}
	if len(summary.CountryBreakdown) != 3 {
		t.Errorf("CountryBreakdown has %d entries, want 3", len(summary.CountryBreakdown))
	}
}
