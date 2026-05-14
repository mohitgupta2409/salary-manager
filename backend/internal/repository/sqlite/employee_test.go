package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// testFixture holds an in-memory DB and seeded reference data for testing.
type testFixture struct {
	repo       *employeeRepo
	countryID  int64
	otherCountryID int64
	jobTitleID int64
	otherTitleID int64
	deptID     int64
	otherDeptID int64
}

func setupFixture(t *testing.T) *testFixture {
	t.Helper()
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	cr := &countryRepo{db: db}
	dr := &departmentRepo{db: db}
	jr := &jobTitleRepo{db: db}

	usa := &model.Country{Name: "United States", Code: "US", Currency: "USD"}
	if err := cr.Create(ctx, usa); err != nil {
		t.Fatalf("seed country: %v", err)
	}
	india := &model.Country{Name: "India", Code: "IN", Currency: "INR"}
	if err := cr.Create(ctx, india); err != nil {
		t.Fatalf("seed country: %v", err)
	}

	eng := &model.Department{Name: "Engineering"}
	if err := dr.Create(ctx, eng); err != nil {
		t.Fatalf("seed dept: %v", err)
	}
	mkt := &model.Department{Name: "Marketing"}
	if err := dr.Create(ctx, mkt); err != nil {
		t.Fatalf("seed dept: %v", err)
	}

	se := &model.JobTitle{Name: "Software Engineer", DepartmentID: eng.ID}
	if err := jr.Create(ctx, se); err != nil {
		t.Fatalf("seed job title: %v", err)
	}
	mm := &model.JobTitle{Name: "Marketing Manager", DepartmentID: mkt.ID}
	if err := jr.Create(ctx, mm); err != nil {
		t.Fatalf("seed job title: %v", err)
	}

	return &testFixture{
		repo:           &employeeRepo{db: db},
		countryID:      usa.ID,
		otherCountryID: india.ID,
		jobTitleID:     se.ID,
		otherTitleID:   mm.ID,
		deptID:         eng.ID,
		otherDeptID:    mkt.ID,
	}
}

func newTestEmployee(f *testFixture, first, last string, salary float64) *model.Employee {
	return &model.Employee{
		FirstName:  first,
		LastName:   last,
		Email:      fmt.Sprintf("%s.%s@test.com", first, last),
		JobTitleID: f.jobTitleID,
		CountryID:  f.countryID,
		Salary:     salary,
		Address:    "123 Main St",
		JoinDate:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreate(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Alice", "Johnson", 120000)
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if emp.ID == 0 {
		t.Error("Create() should set ID")
	}
	if emp.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}
	if !emp.IsActive {
		t.Error("Create() should set IsActive to true")
	}
	// Denormalized fields should be populated
	if emp.Country != "United States" {
		t.Errorf("Country = %q, want United States", emp.Country)
	}
	if emp.Currency != "USD" {
		t.Errorf("Currency = %q, want USD", emp.Currency)
	}
	if emp.JobTitle != "Software Engineer" {
		t.Errorf("JobTitle = %q, want Software Engineer", emp.JobTitle)
	}
	if emp.Department != "Engineering" {
		t.Errorf("Department = %q, want Engineering", emp.Department)
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp1 := newTestEmployee(f, "Alice", "Johnson", 100000)
	if err := f.repo.Create(ctx, emp1); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	emp2 := newTestEmployee(f, "Alice", "Johnson", 120000)
	emp2.Email = emp1.Email
	if err := f.repo.Create(ctx, emp2); err == nil {
		t.Error("Create() with duplicate email should return error")
	}
}

func TestCreate_InvalidForeignKey(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Bob", "Smith", 100000)
	emp.CountryID = 9999 // does not exist
	if err := f.repo.Create(ctx, emp); err == nil {
		t.Error("Create() with bad country FK should return error")
	}
}

func TestGetByID(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Bob", "Smith", 90000)
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := f.repo.GetByID(ctx, emp.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetByID() returned nil")
	}
	if got.FirstName != "Bob" || got.LastName != "Smith" {
		t.Errorf("name = %s %s, want Bob Smith", got.FirstName, got.LastName)
	}
	if got.Country != "United States" {
		t.Errorf("Country = %q, want United States", got.Country)
	}
	if got.Salary != 90000 {
		t.Errorf("Salary = %f, want 90000", got.Salary)
	}
	if got.Address != "123 Main St" {
		t.Errorf("Address = %q, want 123 Main St", got.Address)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	got, err := f.repo.GetByID(ctx, 9999)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got != nil {
		t.Error("GetByID() for non-existent ID should return nil")
	}
}

func TestUpdate(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Carol", "White", 85000)
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	emp.Salary = 95000
	emp.JobTitleID = f.otherTitleID
	if err := f.repo.Update(ctx, emp); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, _ := f.repo.GetByID(ctx, emp.ID)
	if got.Salary != 95000 {
		t.Errorf("after Update(), Salary = %f, want 95000", got.Salary)
	}
	if got.JobTitle != "Marketing Manager" {
		t.Errorf("after Update(), JobTitle = %q, want Marketing Manager", got.JobTitle)
	}
	if got.Department != "Marketing" {
		t.Errorf("after Update(), Department = %q, want Marketing", got.Department)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := &model.Employee{
		ID: 9999, FirstName: "Nobody", LastName: "Ghost",
		Email: "no@one.com", JobTitleID: f.jobTitleID,
		CountryID: f.countryID, Salary: 0, JoinDate: time.Now(),
	}
	if err := f.repo.Update(ctx, emp); err == nil {
		t.Error("Update() for non-existent ID should return error")
	}
}

func TestSoftDelete(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Dave", "Brown", 70000)
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := f.repo.SoftDelete(ctx, emp.ID); err != nil {
		t.Fatalf("SoftDelete() error = %v", err)
	}

	// Row still exists but marked inactive
	got, _ := f.repo.GetByID(ctx, emp.ID)
	if got == nil {
		t.Fatal("GetByID after soft delete should still find the row")
	}
	if got.IsActive {
		t.Error("after SoftDelete(), IsActive should be false")
	}

	// List should not return it
	result, _ := f.repo.List(ctx, model.EmployeeFilter{Page: 1, Limit: 20})
	if result.Total != 0 {
		t.Errorf("List() after soft delete returned %d, want 0", result.Total)
	}
}

func TestSoftDelete_AlreadyDeleted(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Eve", "Black", 80000)
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_ = f.repo.SoftDelete(ctx, emp.ID)

	// Soft-deleting again should fail (row is already inactive)
	if err := f.repo.SoftDelete(ctx, emp.ID); err == nil {
		t.Error("second SoftDelete() should fail (already inactive)")
	}
}

func TestSoftDelete_NotFound(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	if err := f.repo.SoftDelete(ctx, 9999); err == nil {
		t.Error("SoftDelete() for non-existent ID should return error")
	}
}

func TestList_Pagination(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	for i := 0; i < 25; i++ {
		emp := newTestEmployee(f, fmt.Sprintf("Emp%d", i), "Test", float64(50000+i*1000))
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := f.repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := f.repo.List(ctx, model.EmployeeFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 25 {
		t.Errorf("Total = %d, want 25", result.Total)
	}
	if len(result.Employees) != 10 {
		t.Errorf("page 1 returned %d, want 10", len(result.Employees))
	}

	result, err = f.repo.List(ctx, model.EmployeeFilter{Page: 3, Limit: 10})
	if err != nil {
		t.Fatalf("List() page 3 error = %v", err)
	}
	if len(result.Employees) != 5 {
		t.Errorf("page 3 returned %d, want 5", len(result.Employees))
	}
}

func TestList_FilterByCountry(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		emp := newTestEmployee(f, fmt.Sprintf("USA%d", i), "Person", 80000)
		emp.Email = fmt.Sprintf("usa_%d@test.com", i)
		if err := f.repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}
	emp := newTestEmployee(f, "Indian", "Person", 80000)
	emp.Email = "india@test.com"
	emp.CountryID = f.otherCountryID
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	result, err := f.repo.List(ctx, model.EmployeeFilter{CountryID: f.countryID, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 3 {
		t.Errorf("USA total = %d, want 3", result.Total)
	}
}

func TestList_FilterByDepartment(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp1 := newTestEmployee(f, "Eng1", "Person", 80000)
	emp1.Email = "eng1@test.com"
	if err := f.repo.Create(ctx, emp1); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	emp2 := newTestEmployee(f, "Mkt1", "Person", 80000)
	emp2.Email = "mkt1@test.com"
	emp2.JobTitleID = f.otherTitleID
	if err := f.repo.Create(ctx, emp2); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	result, err := f.repo.List(ctx, model.EmployeeFilter{DepartmentID: f.deptID, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Engineering total = %d, want 1", result.Total)
	}
	if result.Employees[0].FirstName != "Eng1" {
		t.Errorf("got %s, want Eng1", result.Employees[0].FirstName)
	}
}

func TestList_FilterBySearch(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	for i, name := range []string{"Alice", "Bob", "Aaron"} {
		emp := newTestEmployee(f, name, "Person", 80000)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := f.repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	result, err := f.repo.List(ctx, model.EmployeeFilter{Search: "A", Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if result.Total != 2 {
		t.Errorf("search 'A' total = %d, want 2 (Alice, Aaron)", result.Total)
	}
}

func TestGetSalaryRangeByCountry(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	salaries := []float64{60000, 80000, 100000, 120000, 140000}
	for i, s := range salaries {
		emp := newTestEmployee(f, fmt.Sprintf("Emp%d", i), "Person", s)
		emp.Email = fmt.Sprintf("emp_%d@test.com", i)
		if err := f.repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create error: %v", err)
		}
	}

	sr, err := f.repo.GetSalaryRangeByCountry(ctx, "United States")
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

func TestGetSalaryRange_IgnoresInactive(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp := newTestEmployee(f, "Active", "Person", 100000)
	if err := f.repo.Create(ctx, emp); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	inactive := newTestEmployee(f, "Inactive", "Person", 999999)
	inactive.Email = "inactive@test.com"
	if err := f.repo.Create(ctx, inactive); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if err := f.repo.SoftDelete(ctx, inactive.ID); err != nil {
		t.Fatalf("SoftDelete error: %v", err)
	}

	sr, err := f.repo.GetSalaryRangeByCountry(ctx, "United States")
	if err != nil {
		t.Fatalf("GetSalaryRangeByCountry() error = %v", err)
	}
	if sr.Count != 1 || sr.Average != 100000 {
		t.Errorf("inactive employee should be excluded; got count=%d avg=%f", sr.Count, sr.Average)
	}
}

func TestGetSalaryByTitle(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	salaries := []float64{100000, 120000}
	for i, s := range salaries {
		emp := newTestEmployee(f, fmt.Sprintf("Eng%d", i), "Person", s)
		emp.Email = fmt.Sprintf("eng_%d@test.com", i)
		if err := f.repo.Create(ctx, emp); err != nil {
			t.Fatalf("Create error: %v", err)
		}
	}
	mkt := newTestEmployee(f, "Mkt", "Person", 150000)
	mkt.Email = "mkt@test.com"
	mkt.JobTitleID = f.otherTitleID
	if err := f.repo.Create(ctx, mkt); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	sbt, err := f.repo.GetSalaryByTitle(ctx, "United States", "Software Engineer")
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
	f := setupFixture(t)
	ctx := context.Background()

	emp1 := newTestEmployee(f, "Eng", "One", 100000)
	emp1.Email = "eng_one@test.com"
	if err := f.repo.Create(ctx, emp1); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	emp2 := newTestEmployee(f, "Eng", "Two", 120000)
	emp2.Email = "eng_two@test.com"
	if err := f.repo.Create(ctx, emp2); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	mkt := newTestEmployee(f, "Mkt", "One", 80000)
	mkt.Email = "mkt_one@test.com"
	mkt.JobTitleID = f.otherTitleID
	if err := f.repo.Create(ctx, mkt); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	stats, err := f.repo.GetDepartmentStats(ctx, "United States")
	if err != nil {
		t.Fatalf("GetDepartmentStats() error = %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("got %d departments, want 2", len(stats))
	}
	// Engineering should be first (110k avg vs 80k)
	if stats[0].Department != "Engineering" {
		t.Errorf("first = %q, want Engineering", stats[0].Department)
	}
	if stats[0].EmployeeCount != 2 {
		t.Errorf("Engineering count = %d, want 2", stats[0].EmployeeCount)
	}
}

func TestGetOrgSummary(t *testing.T) {
	f := setupFixture(t)
	ctx := context.Background()

	emp1 := newTestEmployee(f, "USA", "One", 100000)
	emp1.Email = "usa1@test.com"
	if err := f.repo.Create(ctx, emp1); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	emp2 := newTestEmployee(f, "USA", "Two", 100000)
	emp2.Email = "usa2@test.com"
	if err := f.repo.Create(ctx, emp2); err != nil {
		t.Fatalf("Create error: %v", err)
	}
	india := newTestEmployee(f, "Indian", "Person", 100000)
	india.Email = "india@test.com"
	india.CountryID = f.otherCountryID
	if err := f.repo.Create(ctx, india); err != nil {
		t.Fatalf("Create error: %v", err)
	}

	summary, err := f.repo.GetOrgSummary(ctx)
	if err != nil {
		t.Fatalf("GetOrgSummary() error = %v", err)
	}
	if summary.TotalEmployees != 3 {
		t.Errorf("TotalEmployees = %d, want 3", summary.TotalEmployees)
	}
	if summary.TotalCountries != 2 {
		t.Errorf("TotalCountries = %d, want 2", summary.TotalCountries)
	}
	if len(summary.CountryBreakdown) != 2 {
		t.Errorf("CountryBreakdown has %d entries, want 2", len(summary.CountryBreakdown))
	}
}
