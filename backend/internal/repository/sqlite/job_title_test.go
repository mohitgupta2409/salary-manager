package sqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

// jobTitleFixture seeds a couple of departments and returns a configured
// jobTitleRepo so each test can focus on the JobTitleRepository under test.
type jobTitleFixture struct {
	db        *sql.DB
	repo      *jobTitleRepo
	engID     int64
	salesID   int64
	mktID     int64
}

func newJobTitleFixture(t *testing.T) *jobTitleFixture {
	t.Helper()
	db := newTestDB(t)
	dr := &departmentRepo{db: db}
	ctx := context.Background()

	eng := &model.Department{Name: "Engineering"}
	sales := &model.Department{Name: "Sales"}
	mkt := &model.Department{Name: "Marketing"}
	for _, d := range []*model.Department{eng, sales, mkt} {
		if err := dr.Create(ctx, d); err != nil {
			t.Fatalf("seed dept %s: %v", d.Name, err)
		}
	}
	return &jobTitleFixture{
		db:      db,
		repo:    &jobTitleRepo{db: db},
		engID:   eng.ID,
		salesID: sales.ID,
		mktID:   mkt.ID,
	}
}

// ----- Create -----

func TestJobTitle_Create_Success(t *testing.T) {
	f := newJobTitleFixture(t)
	jt := &model.JobTitle{Name: "Software Engineer", DepartmentID: f.engID}

	if err := f.repo.Create(context.Background(), jt); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if jt.ID == 0 {
		t.Error("Create() should assign an ID")
	}
	if !jt.IsActive {
		t.Error("Create() should default IsActive to true")
	}
	if jt.CreatedAt.IsZero() || jt.UpdatedAt.IsZero() {
		t.Error("Create() should set timestamps")
	}
}

func TestJobTitle_Create_DuplicateInSameDepartment(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	if err := f.repo.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: f.engID}); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	err := f.repo.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: f.engID})
	if err == nil {
		t.Error("expected duplicate (name, department_id) insert to fail")
	}
}

func TestJobTitle_Create_SameNameDifferentDepartments(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	if err := f.repo.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: f.engID}); err != nil {
		t.Fatalf("eng Manager: %v", err)
	}
	if err := f.repo.Create(ctx, &model.JobTitle{Name: "Manager", DepartmentID: f.salesID}); err != nil {
		t.Errorf("Manager in a different department should be allowed: %v", err)
	}
}

func TestJobTitle_Create_InvalidDepartmentFK(t *testing.T) {
	f := newJobTitleFixture(t)
	err := f.repo.Create(context.Background(), &model.JobTitle{Name: "Ghost", DepartmentID: 9999})
	if err == nil {
		t.Error("expected FK violation when department_id does not exist")
	}
}

// ----- GetByID -----

func TestJobTitle_GetByID_PopulatesDepartmentName(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	jt := &model.JobTitle{Name: "Software Engineer", DepartmentID: f.engID}
	if err := f.repo.Create(ctx, jt); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := f.repo.GetByID(ctx, jt.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil for an existing id")
	}
	if got.Name != "Software Engineer" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.DepartmentID != f.engID {
		t.Errorf("DepartmentID = %d, want %d", got.DepartmentID, f.engID)
	}
	if got.Department != "Engineering" {
		t.Errorf("Department (denormalized) = %q, want Engineering", got.Department)
	}
}

func TestJobTitle_GetByID_NotFound(t *testing.T) {
	f := newJobTitleFixture(t)
	got, err := f.repo.GetByID(context.Background(), 9999)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got != nil {
		t.Errorf("missing id should return nil, got %+v", got)
	}
}

// ----- List -----

func TestJobTitle_List_EmptyDatabase(t *testing.T) {
	f := newJobTitleFixture(t)
	out, err := f.repo.List(context.Background(), model.JobTitleListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 0 {
		t.Errorf("Total = %d, want 0", out.Total)
	}
	if out.JobTitles == nil || len(out.JobTitles) != 0 {
		t.Errorf("expected empty slice, got %v", out.JobTitles)
	}
}

func TestJobTitle_List_AllRecords_NoLimit(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	for _, jt := range []*model.JobTitle{
		{Name: "Software Engineer", DepartmentID: f.engID},
		{Name: "DevOps Engineer", DepartmentID: f.engID},
		{Name: "Account Executive", DepartmentID: f.salesID},
	} {
		if err := f.repo.Create(ctx, jt); err != nil {
			t.Fatalf("Create %s: %v", jt.Name, err)
		}
	}

	out, err := f.repo.List(ctx, model.JobTitleListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 3 || len(out.JobTitles) != 3 {
		t.Errorf("Total=%d returned=%d, want 3 of each", out.Total, len(out.JobTitles))
	}
}

func TestJobTitle_List_FilterByDepartment(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	for _, jt := range []*model.JobTitle{
		{Name: "Software Engineer", DepartmentID: f.engID},
		{Name: "Senior Engineer", DepartmentID: f.engID},
		{Name: "Marketing Manager", DepartmentID: f.mktID},
	} {
		if err := f.repo.Create(ctx, jt); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	out, err := f.repo.List(ctx, model.JobTitleListRequest{DepartmentID: f.engID})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 2 {
		t.Errorf("Total = %d, want 2 (Engineering only)", out.Total)
	}
	for _, jt := range out.JobTitles {
		if jt.DepartmentID != f.engID {
			t.Errorf("got jt with DepartmentID=%d, want %d", jt.DepartmentID, f.engID)
		}
		if jt.Department != "Engineering" {
			t.Errorf("Department (denormalized) = %q, want Engineering", jt.Department)
		}
	}
}

func TestJobTitle_List_OrderedByDeptThenName(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	// Three departments: Engineering, Marketing, Sales (alphabetical)
	for _, jt := range []*model.JobTitle{
		{Name: "Z-Engineer", DepartmentID: f.engID},
		{Name: "A-Engineer", DepartmentID: f.engID},
		{Name: "Marketer", DepartmentID: f.mktID},
		{Name: "Salesperson", DepartmentID: f.salesID},
	} {
		if err := f.repo.Create(ctx, jt); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	out, err := f.repo.List(ctx, model.JobTitleListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []string{"A-Engineer", "Z-Engineer", "Marketer", "Salesperson"}
	for i, jt := range out.JobTitles {
		if jt.Name != want[i] {
			t.Errorf("position %d = %q, want %q", i, jt.Name, want[i])
		}
	}
}

func TestJobTitle_List_PaginationWindow(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	for _, name := range []string{"Engineer 1", "Engineer 2", "Engineer 3", "Engineer 4", "Engineer 5"} {
		if err := f.repo.Create(ctx, &model.JobTitle{Name: name, DepartmentID: f.engID}); err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
	}

	out, err := f.repo.List(ctx, model.JobTitleListRequest{
		DepartmentID: f.engID,
		Pagination:   model.Pagination{Limit: 2, Offset: 1},
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 5 {
		t.Errorf("Total = %d, want 5", out.Total)
	}
	if len(out.JobTitles) != 2 {
		t.Errorf("page size = %d, want 2", len(out.JobTitles))
	}
	if out.Limit != 2 || out.Offset != 1 {
		t.Errorf("echoed pagination = limit=%d offset=%d, want 2/1", out.Limit, out.Offset)
	}
}

func TestJobTitle_List_ExcludesInactiveByDefault(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	active := &model.JobTitle{Name: "Active Role", DepartmentID: f.engID}
	inactive := &model.JobTitle{Name: "Old Role", DepartmentID: f.engID}
	for _, jt := range []*model.JobTitle{active, inactive} {
		if err := f.repo.Create(ctx, jt); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	deactivateJobTitle(t, f.db, inactive.ID)

	out, err := f.repo.List(ctx, model.JobTitleListRequest{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 || out.JobTitles[0].Name != "Active Role" {
		t.Errorf("active filter failed: total=%d titles=%v", out.Total, out.JobTitles)
	}
}

func TestJobTitle_List_IncludeInactive(t *testing.T) {
	f := newJobTitleFixture(t)
	ctx := context.Background()
	a := &model.JobTitle{Name: "Active", DepartmentID: f.engID}
	b := &model.JobTitle{Name: "Old", DepartmentID: f.engID}
	for _, jt := range []*model.JobTitle{a, b} {
		if err := f.repo.Create(ctx, jt); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	deactivateJobTitle(t, f.db, b.ID)

	out, err := f.repo.List(ctx, model.JobTitleListRequest{IncludeInactive: true})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 2 || len(out.JobTitles) != 2 {
		t.Errorf("IncludeInactive=true should return both, got total=%d returned=%d",
			out.Total, len(out.JobTitles))
	}
}
