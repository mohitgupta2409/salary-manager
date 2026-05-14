package sqlite

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

func setupRefDB(t *testing.T) (*countryRepo, *departmentRepo, *jobTitleRepo) {
	t.Helper()
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &countryRepo{db: db}, &departmentRepo{db: db}, &jobTitleRepo{db: db}
}

func TestCountry_CreateAndGet(t *testing.T) {
	cr, _, _ := setupRefDB(t)
	ctx := context.Background()

	c := &model.Country{Name: "United States", Code: "us", Currency: "usd"}
	if err := cr.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.ID == 0 {
		t.Error("Create() should set ID")
	}
	if c.Code != "US" {
		t.Errorf("Code = %q, want uppercase US", c.Code)
	}
	if c.Currency != "USD" {
		t.Errorf("Currency = %q, want uppercase USD", c.Currency)
	}
	if !c.IsActive {
		t.Error("IsActive should default to true")
	}

	got, err := cr.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Name != "United States" {
		t.Errorf("Name = %q", got.Name)
	}
}

func TestCountry_GetByCode(t *testing.T) {
	cr, _, _ := setupRefDB(t)
	ctx := context.Background()

	c := &model.Country{Name: "India", Code: "IN", Currency: "INR"}
	if err := cr.Create(ctx, c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := cr.GetByCode(ctx, "in") // case-insensitive
	if err != nil {
		t.Fatalf("GetByCode() error = %v", err)
	}
	if got == nil || got.Name != "India" {
		t.Errorf("GetByCode('in') did not find India")
	}
}

func TestCountry_DuplicateCode(t *testing.T) {
	cr, _, _ := setupRefDB(t)
	ctx := context.Background()

	c1 := &model.Country{Name: "United States", Code: "US", Currency: "USD"}
	if err := cr.Create(ctx, c1); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	c2 := &model.Country{Name: "Different Name", Code: "US", Currency: "USD"}
	if err := cr.Create(ctx, c2); err == nil {
		t.Error("Create() with duplicate code should fail")
	}
}

func TestCountry_List(t *testing.T) {
	cr, _, _ := setupRefDB(t)
	ctx := context.Background()

	for _, c := range []model.Country{
		{Name: "Zambia", Code: "ZM", Currency: "ZMW"},
		{Name: "Argentina", Code: "AR", Currency: "ARS"},
	} {
		c := c
		if err := cr.Create(ctx, &c); err != nil {
			t.Fatalf("Create error: %v", err)
		}
	}

	out, err := cr.List(ctx, false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("List() returned %d, want 2", len(out))
	}
	// Should be sorted by name (Argentina before Zambia)
	if out[0].Name != "Argentina" {
		t.Errorf("first = %q, want Argentina", out[0].Name)
	}
}

func TestDepartment_CreateAndList(t *testing.T) {
	_, dr, _ := setupRefDB(t)
	ctx := context.Background()

	d := &model.Department{Name: "Engineering"}
	if err := dr.Create(ctx, d); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if d.ID == 0 || !d.IsActive {
		t.Error("Create() should set ID and IsActive")
	}

	out, err := dr.List(ctx, false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(out) != 1 {
		t.Errorf("List() returned %d, want 1", len(out))
	}
}

func TestDepartment_DuplicateName(t *testing.T) {
	_, dr, _ := setupRefDB(t)
	ctx := context.Background()

	if err := dr.Create(ctx, &model.Department{Name: "Engineering"}); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if err := dr.Create(ctx, &model.Department{Name: "Engineering"}); err == nil {
		t.Error("Create() with duplicate name should fail")
	}
}

func TestJobTitle_CreateAndList(t *testing.T) {
	_, dr, jr := setupRefDB(t)
	ctx := context.Background()

	eng := &model.Department{Name: "Engineering"}
	if err := dr.Create(ctx, eng); err != nil {
		t.Fatalf("seed dept: %v", err)
	}

	jt := &model.JobTitle{Name: "Software Engineer", DepartmentID: eng.ID}
	if err := jr.Create(ctx, jt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if jt.ID == 0 {
		t.Error("Create() should set ID")
	}

	out, err := jr.List(ctx, 0, false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("got %d, want 1", len(out))
	}
	if out[0].Department != "Engineering" {
		t.Errorf("Department = %q, want Engineering", out[0].Department)
	}
}

func TestJobTitle_FilterByDepartment(t *testing.T) {
	_, dr, jr := setupRefDB(t)
	ctx := context.Background()

	eng := &model.Department{Name: "Engineering"}
	mkt := &model.Department{Name: "Marketing"}
	for _, d := range []*model.Department{eng, mkt} {
		if err := dr.Create(ctx, d); err != nil {
			t.Fatalf("seed dept: %v", err)
		}
	}

	for _, jt := range []*model.JobTitle{
		{Name: "Software Engineer", DepartmentID: eng.ID},
		{Name: "Senior Engineer", DepartmentID: eng.ID},
		{Name: "Marketing Manager", DepartmentID: mkt.ID},
	} {
		if err := jr.Create(ctx, jt); err != nil {
			t.Fatalf("Create error: %v", err)
		}
	}

	engTitles, err := jr.List(ctx, eng.ID, false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(engTitles) != 2 {
		t.Errorf("Engineering titles = %d, want 2", len(engTitles))
	}
}

func TestJobTitle_DuplicateInSameDepartment(t *testing.T) {
	_, dr, jr := setupRefDB(t)
	ctx := context.Background()

	eng := &model.Department{Name: "Engineering"}
	if err := dr.Create(ctx, eng); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := jr.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: eng.ID}); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	if err := jr.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: eng.ID}); err == nil {
		t.Error("Create() with duplicate (name, dept) should fail")
	}
}

func TestJobTitle_SameNameDifferentDepartments(t *testing.T) {
	_, dr, jr := setupRefDB(t)
	ctx := context.Background()

	eng := &model.Department{Name: "Engineering"}
	sales := &model.Department{Name: "Sales"}
	for _, d := range []*model.Department{eng, sales} {
		if err := dr.Create(ctx, d); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	if err := jr.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: eng.ID}); err != nil {
		t.Fatalf("eng Manager error = %v", err)
	}
	if err := jr.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: sales.ID}); err != nil {
		t.Errorf("sales Manager should be allowed: %v", err)
	}
}
