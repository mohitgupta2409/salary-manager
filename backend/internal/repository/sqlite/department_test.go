package sqlite

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

func newDepartmentRepoForTest(t *testing.T) *departmentRepo {
	t.Helper()
	return &departmentRepo{db: newTestDB(t)}
}

// ----- Create -----

func TestDepartment_Create_Success(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()

	d := &model.Department{Name: "Engineering"}
	if err := dr.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.ID == 0 {
		t.Error("Create() should assign an ID")
	}
	if !d.IsActive {
		t.Error("Create() should default IsActive to true")
	}
	if d.CreatedAt.IsZero() || d.UpdatedAt.IsZero() {
		t.Error("Create() should set timestamps")
	}
}

func TestDepartment_Create_DuplicateName(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	if err := dr.Create(ctx, &model.Department{Name: "Engineering"}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	if err := dr.Create(ctx, &model.Department{Name: "Engineering"}); err == nil {
		t.Error("expected duplicate-name insert to fail")
	}
}

// ----- GetByID -----

func TestDepartment_GetByID_Success(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	d := &model.Department{Name: "Marketing"}
	if err := dr.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := dr.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil || got.Name != "Marketing" {
		t.Errorf("got %+v, want Marketing", got)
	}
}

func TestDepartment_GetByID_NotFound(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	got, err := dr.GetByID(context.Background(), 9999)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got != nil {
		t.Errorf("missing id should return nil, got %+v", got)
	}
}

// ----- List -----

func TestDepartment_List_EmptyDatabase(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	out, err := dr.List(context.Background(), model.DepartmentListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 0 {
		t.Errorf("Total = %d, want 0", out.Total)
	}
	if out.Departments == nil || len(out.Departments) != 0 {
		t.Errorf("expected empty slice, got %v", out.Departments)
	}
}

func TestDepartment_List_AllRecords_NoLimit(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	for _, name := range []string{"Sales", "Engineering", "Finance"} {
		if err := dr.Create(ctx, &model.Department{Name: name}); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	out, err := dr.List(ctx, model.DepartmentListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 3 || len(out.Departments) != 3 {
		t.Errorf("Total=%d returned=%d, want 3 of each", out.Total, len(out.Departments))
	}
}

func TestDepartment_List_SortedByName(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	for _, name := range []string{"Sales", "Engineering", "Finance"} {
		if err := dr.Create(ctx, &model.Department{Name: name}); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	out, err := dr.List(ctx, model.DepartmentListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []string{"Engineering", "Finance", "Sales"}
	for i, d := range out.Departments {
		if d.Name != want[i] {
			t.Errorf("Departments[%d] = %q, want %q", i, d.Name, want[i])
		}
	}
}

func TestDepartment_List_PaginationWindow(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	for _, name := range []string{"Engineering", "Finance", "HR", "IT", "Legal"} {
		if err := dr.Create(ctx, &model.Department{Name: name}); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	out, err := dr.List(ctx, model.DepartmentListRequest{
		Pagination: model.Pagination{Limit: 2, Offset: 2},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 5 {
		t.Errorf("Total = %d, want 5", out.Total)
	}
	if len(out.Departments) != 2 {
		t.Fatalf("page size = %d, want 2", len(out.Departments))
	}
	// Sorted: Engineering, Finance, HR, IT, Legal -> offset 2, limit 2 -> HR, IT
	if out.Departments[0].Name != "HR" || out.Departments[1].Name != "IT" {
		t.Errorf("got %v, want [HR IT]", []string{out.Departments[0].Name, out.Departments[1].Name})
	}
}

func TestDepartment_List_ExcludesInactiveByDefault(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	active := &model.Department{Name: "Active"}
	inactive := &model.Department{Name: "Old"}
	for _, d := range []*model.Department{active, inactive} {
		if err := dr.Create(ctx, d); err != nil {
			t.Fatalf("Create %s: %v", d.Name, err)
		}
	}
	deactivateDepartment(t, dr.db, inactive.ID)

	out, err := dr.List(ctx, model.DepartmentListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 || out.Departments[0].Name != "Active" {
		t.Errorf("active filter failed: total=%d departments=%v", out.Total, out.Departments)
	}
}

func TestDepartment_List_IncludeInactive(t *testing.T) {
	dr := newDepartmentRepoForTest(t)
	ctx := context.Background()
	a := &model.Department{Name: "Active"}
	b := &model.Department{Name: "Old"}
	for _, d := range []*model.Department{a, b} {
		if err := dr.Create(ctx, d); err != nil {
			t.Fatalf("Create %s: %v", d.Name, err)
		}
	}
	deactivateDepartment(t, dr.db, b.ID)

	out, err := dr.List(ctx, model.DepartmentListRequest{IncludeInactive: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 2 || len(out.Departments) != 2 {
		t.Errorf("IncludeInactive=true should return both rows, got total=%d returned=%d",
			out.Total, len(out.Departments))
	}
}
