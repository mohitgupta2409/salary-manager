package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// =============================================================================
// Create
// =============================================================================

func TestEmployee_Create_Success_PopulatesDenormalizedFields(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()

	e := f.makeEmployee("Alice", "Johnson", 120000)
	if err := f.repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if e.ID == 0 {
		t.Error("Create() should assign an ID")
	}
	if !e.IsActive {
		t.Error("Create() should default IsActive to true")
	}
	if e.CreatedAt.IsZero() || e.UpdatedAt.IsZero() {
		t.Error("Create() should set timestamps")
	}
	// Denormalized fields populated via JOIN re-fetch
	if e.Country != "United States" || e.Currency != "USD" {
		t.Errorf("Country/Currency = %q/%q, want United States/USD", e.Country, e.Currency)
	}
	if e.JobTitle != "Software Engineer" || e.Department != "Engineering" {
		t.Errorf("JobTitle/Department = %q/%q, want Software Engineer/Engineering", e.JobTitle, e.Department)
	}
}

func TestEmployee_Create_EmptyAddress_StoredAsNull(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("NoAddr", "Person", 50000)
	e.Address = ""

	if err := f.repo.Create(ctx, e); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := f.repo.GetByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Address != "" {
		t.Errorf("expected empty Address (NULL in DB), got %q", got.Address)
	}
}

func TestEmployee_Create_DuplicateEmail(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e1 := f.makeEmployee("Alice", "Johnson", 100000)
	if err := f.repo.Create(ctx, e1); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	e2 := f.makeEmployee("Alice", "Johnson", 110000) // same email by construction
	if err := f.repo.Create(ctx, e2); err == nil {
		t.Error("expected duplicate-email insert to fail")
	}
}

func TestEmployee_Create_InvalidCountryFK(t *testing.T) {
	f := newEmployeeFixture(t)
	e := f.makeEmployee("Bob", "BadCountry", 80000)
	e.CountryID = 9999
	if err := f.repo.Create(context.Background(), e); err == nil {
		t.Error("expected FK violation for unknown country_id")
	}
}

func TestEmployee_Create_InvalidJobTitleFK(t *testing.T) {
	f := newEmployeeFixture(t)
	e := f.makeEmployee("Bob", "BadTitle", 80000)
	e.JobTitleID = 9999
	if err := f.repo.Create(context.Background(), e); err == nil {
		t.Error("expected FK violation for unknown job_title_id")
	}
}

func TestEmployee_Create_NegativeSalary_RejectedByCheck(t *testing.T) {
	f := newEmployeeFixture(t)
	e := f.makeEmployee("Bob", "Negative", -1)
	if err := f.repo.Create(context.Background(), e); err == nil {
		t.Error("expected CHECK(salary >= 0) to reject negative salary")
	}
}

// =============================================================================
// GetByID
// =============================================================================

func TestEmployee_GetByID_Success(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("Bob", "Smith", 90000)
	f.mustCreateEmployee(t, e)

	got, err := f.repo.GetByID(ctx, e.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil for an existing id")
	}
	if got.FirstName != "Bob" || got.LastName != "Smith" {
		t.Errorf("name = %s %s, want Bob Smith", got.FirstName, got.LastName)
	}
	if got.Salary != 90000 {
		t.Errorf("Salary = %f, want 90000", got.Salary)
	}
	if got.Address != "1 Test Way" {
		t.Errorf("Address = %q, want 1 Test Way", got.Address)
	}
	if got.Country != "United States" {
		t.Errorf("Country (denormalized) = %q", got.Country)
	}
}

func TestEmployee_GetByID_NotFound(t *testing.T) {
	f := newEmployeeFixture(t)
	got, err := f.repo.GetByID(context.Background(), 9999)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got != nil {
		t.Errorf("missing id should return nil, got %+v", got)
	}
}

// =============================================================================
// Update
// =============================================================================

func TestEmployee_Update_PersistsChanges_AndRefreshesDenormalized(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("Carol", "White", 85000)
	f.mustCreateEmployee(t, e)

	originalUpdatedAt := e.UpdatedAt
	time.Sleep(2 * time.Millisecond) // ensure UpdatedAt changes

	e.Salary = 95000
	e.JobTitleID = f.mktMgrTitleID // moves to Marketing
	if err := f.repo.Update(ctx, e); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := f.repo.GetByID(ctx, e.ID)
	if got.Salary != 95000 {
		t.Errorf("Salary = %f, want 95000", got.Salary)
	}
	if got.JobTitle != "Marketing Manager" {
		t.Errorf("JobTitle = %q, want Marketing Manager", got.JobTitle)
	}
	if got.Department != "Marketing" {
		t.Errorf("Department = %q, want Marketing", got.Department)
	}
	if !got.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("UpdatedAt should advance: before=%v after=%v", originalUpdatedAt, got.UpdatedAt)
	}
}

func TestEmployee_Update_ClearsAddress(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("Carol", "Address", 80000)
	f.mustCreateEmployee(t, e)

	e.Address = ""
	if err := f.repo.Update(ctx, e); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ := f.repo.GetByID(ctx, e.ID)
	if got.Address != "" {
		t.Errorf("Address = %q, want empty after update", got.Address)
	}
}

func TestEmployee_Update_NotFound(t *testing.T) {
	f := newEmployeeFixture(t)
	e := f.makeEmployee("Ghost", "Person", 0)
	e.ID = 9999
	err := f.repo.Update(context.Background(), e)
	if err == nil {
		t.Error("expected Update() to fail for a non-existent id")
	}
}

func TestEmployee_Update_DuplicateEmailFails(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	a := f.makeEmployee("Alice", "First", 100000)
	b := f.makeEmployee("Bob", "Second", 110000)
	f.mustCreateEmployee(t, a)
	f.mustCreateEmployee(t, b)

	b.Email = a.Email
	if err := f.repo.Update(ctx, b); err == nil {
		t.Error("expected unique-email constraint to reject the update")
	}
}

// =============================================================================
// Delete (soft delete)
// =============================================================================

func TestEmployee_Delete_MarksRowInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("Dave", "Brown", 70000)
	f.mustCreateEmployee(t, e)

	if err := f.repo.Delete(ctx, e.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Row still retrievable but inactive
	got, _ := f.repo.GetByID(ctx, e.ID)
	if got == nil {
		t.Fatal("after Delete, GetByID should still return the row")
	}
	if got.IsActive {
		t.Error("after Delete, IsActive should be false")
	}
}

func TestEmployee_Delete_HiddenFromList(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("Hidden", "Person", 70000)
	f.mustCreateEmployee(t, e)

	if err := f.repo.Delete(ctx, e.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	out, err := f.repo.List(ctx, model.EmployeeFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 0 {
		t.Errorf("List.Total = %d, want 0 (deleted employee should not appear)", out.Total)
	}
}

func TestEmployee_Delete_AlreadyInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e := f.makeEmployee("Twice", "Deleted", 80000)
	f.mustCreateEmployee(t, e)
	if err := f.repo.Delete(ctx, e.ID); err != nil {
		t.Fatalf("first Delete: %v", err)
	}
	if err := f.repo.Delete(ctx, e.ID); err == nil {
		t.Error("expected second Delete on already-inactive row to fail")
	}
}

func TestEmployee_Delete_NotFound(t *testing.T) {
	f := newEmployeeFixture(t)
	if err := f.repo.Delete(context.Background(), 9999); err == nil {
		t.Error("expected Delete() of non-existent id to fail")
	}
}

// =============================================================================
// List — pagination & filters
// =============================================================================

func TestEmployee_List_AllRecords_NoLimit(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i := 0; i < 7; i++ {
		e := f.makeEmployee(fmt.Sprintf("Emp%d", i), "Person", float64(50000+i*1000))
		e.Email = fmt.Sprintf("emp_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}

	out, err := f.repo.List(ctx, model.EmployeeFilter{}) // Limit=0 -> all
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 7 || len(out.Employees) != 7 {
		t.Errorf("Total=%d returned=%d, want 7 of each", out.Total, len(out.Employees))
	}
	if out.Limit != 0 || out.Offset != 0 {
		t.Errorf("echoed pagination = limit=%d offset=%d, want 0/0", out.Limit, out.Offset)
	}
}

func TestEmployee_List_PaginationWindow(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i := 0; i < 25; i++ {
		e := f.makeEmployee(fmt.Sprintf("E%02d", i), "Person", float64(50000+i*1000))
		e.Email = fmt.Sprintf("e_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}

	// Page 1 (offset 0, limit 10)
	page1, err := f.repo.List(ctx, model.EmployeeFilter{
		Pagination: model.Pagination{Limit: 10, Offset: 0},
	})
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if page1.Total != 25 {
		t.Errorf("Total = %d, want 25", page1.Total)
	}
	if len(page1.Employees) != 10 {
		t.Errorf("page 1 size = %d, want 10", len(page1.Employees))
	}

	// Last page (offset 20, limit 10) -> only 5 left
	last, err := f.repo.List(ctx, model.EmployeeFilter{
		Pagination: model.Pagination{Limit: 10, Offset: 20},
	})
	if err != nil {
		t.Fatalf("List last page: %v", err)
	}
	if len(last.Employees) != 5 {
		t.Errorf("last page size = %d, want 5", len(last.Employees))
	}
}

func TestEmployee_List_OffsetWithoutLimit_SkipsThenReturnsRest(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		e := f.makeEmployee(fmt.Sprintf("E%d", i), "Person", 80000)
		e.Email = fmt.Sprintf("e_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}

	out, err := f.repo.List(ctx, model.EmployeeFilter{
		Pagination: model.Pagination{Offset: 2}, // Limit=0 (all) but skip 2
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 5 {
		t.Errorf("Total = %d, want 5", out.Total)
	}
	if len(out.Employees) != 3 {
		t.Errorf("after offset 2, returned=%d, want 3", len(out.Employees))
	}
}

func TestEmployee_List_OrderedByIDDesc(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	var ids []int64
	for i := 0; i < 3; i++ {
		e := f.makeEmployee(fmt.Sprintf("E%d", i), "Person", 80000)
		e.Email = fmt.Sprintf("e_%d@example.com", i)
		f.mustCreateEmployee(t, e)
		ids = append(ids, e.ID)
	}

	out, _ := f.repo.List(ctx, model.EmployeeFilter{})
	if out.Employees[0].ID != ids[2] || out.Employees[2].ID != ids[0] {
		t.Errorf("expected DESC order; got %v",
			[]int64{out.Employees[0].ID, out.Employees[1].ID, out.Employees[2].ID})
	}
}

func TestEmployee_List_FilterByCountry(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		e := f.makeEmployee(fmt.Sprintf("USA%d", i), "Person", 80000)
		e.Email = fmt.Sprintf("usa_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}
	indian := f.makeEmployee("Indian", "Person", 80000)
	indian.Email = "india@example.com"
	indian.CountryID = f.indiaID
	f.mustCreateEmployee(t, indian)

	out, err := f.repo.List(ctx, model.EmployeeFilter{CountryID: f.usaID})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 3 {
		t.Errorf("Total = %d, want 3 (USA only)", out.Total)
	}
}

func TestEmployee_List_FilterByJobTitle(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	eng := f.makeEmployee("Eng", "One", 100000)
	eng.Email = "eng@example.com"
	f.mustCreateEmployee(t, eng)
	mkt := f.makeEmployee("Mkt", "One", 95000)
	mkt.Email = "mkt@example.com"
	mkt.JobTitleID = f.mktMgrTitleID
	f.mustCreateEmployee(t, mkt)

	out, err := f.repo.List(ctx, model.EmployeeFilter{JobTitleID: f.mktMgrTitleID})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 || out.Employees[0].FirstName != "Mkt" {
		t.Errorf("expected only the Marketing Manager; got total=%d employees=%v", out.Total, out.Employees)
	}
}

func TestEmployee_List_FilterByDepartment(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	eng := f.makeEmployee("Eng1", "Person", 80000)
	eng.Email = "eng1@example.com"
	f.mustCreateEmployee(t, eng)
	mkt := f.makeEmployee("Mkt1", "Person", 80000)
	mkt.Email = "mkt1@example.com"
	mkt.JobTitleID = f.mktMgrTitleID
	f.mustCreateEmployee(t, mkt)

	out, err := f.repo.List(ctx, model.EmployeeFilter{DepartmentID: f.engID})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 1 || out.Employees[0].FirstName != "Eng1" {
		t.Errorf("expected only the Engineering employee; got %v", out.Employees)
	}
}

func TestEmployee_List_FilterBySearch(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	// Use a search domain free of the search letter so the LIKE only matches
	// the first/last name fields (SQLite LIKE is case-insensitive for ASCII).
	for i, name := range []string{"Alice", "Bob", "Aaron"} {
		e := f.makeEmployee(name, "Person", 80000)
		e.Email = fmt.Sprintf("user_%d@hr.io", i)
		f.mustCreateEmployee(t, e)
	}

	out, err := f.repo.List(ctx, model.EmployeeFilter{Search: "A"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if out.Total != 2 {
		t.Errorf("search 'A' total = %d, want 2 (Alice, Aaron)", out.Total)
	}
}

func TestEmployee_List_SearchMatchesEmail(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	e1 := f.makeEmployee("Anna", "Smith", 100000)
	e1.Email = "anna.smith@example.com"
	f.mustCreateEmployee(t, e1)
	e2 := f.makeEmployee("Beth", "Jones", 100000)
	e2.Email = "different@elsewhere.com"
	f.mustCreateEmployee(t, e2)

	out, _ := f.repo.List(ctx, model.EmployeeFilter{Search: "elsewhere"})
	if out.Total != 1 || out.Employees[0].FirstName != "Beth" {
		t.Errorf("expected to match by email substring; got total=%d employees=%v", out.Total, out.Employees)
	}
}

func TestEmployee_List_FiltersCombine(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	keep := f.makeEmployee("Anna", "Match", 100000)
	keep.Email = "anna.match@example.com"
	f.mustCreateEmployee(t, keep)
	wrongCountry := f.makeEmployee("Anna", "Other", 100000)
	wrongCountry.Email = "anna.other@example.com"
	wrongCountry.CountryID = f.indiaID
	f.mustCreateEmployee(t, wrongCountry)

	out, _ := f.repo.List(ctx, model.EmployeeFilter{
		Search:    "Anna",
		CountryID: f.usaID,
	})
	if out.Total != 1 || out.Employees[0].LastName != "Match" {
		t.Errorf("expected combined filter to keep only the USA Anna; got %v", out.Employees)
	}
}

func TestEmployee_List_ExcludesInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	keep := f.makeEmployee("Active", "Keep", 80000)
	f.mustCreateEmployee(t, keep)
	gone := f.makeEmployee("Inactive", "Gone", 80000)
	f.mustCreateEmployee(t, gone)
	if err := f.repo.Delete(ctx, gone.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	out, _ := f.repo.List(ctx, model.EmployeeFilter{})
	if out.Total != 1 || out.Employees[0].FirstName != "Active" {
		t.Errorf("inactive should be excluded; got %v", out.Employees)
	}
}

// =============================================================================
// Salary insights
// =============================================================================

func TestEmployee_GetSalaryRangeByCountry_ComputesStats(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i, s := range []float64{60000, 80000, 100000, 120000, 140000} {
		e := f.makeEmployee(fmt.Sprintf("E%d", i), "Person", s)
		e.Email = fmt.Sprintf("e_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}

	sr, err := f.repo.GetSalaryRangeByCountry(ctx, "United States")
	if err != nil {
		t.Fatalf("GetSalaryRangeByCountry: %v", err)
	}
	if sr.Min != 60000 || sr.Max != 140000 {
		t.Errorf("Min/Max = %f/%f, want 60000/140000", sr.Min, sr.Max)
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

func TestEmployee_GetSalaryRangeByCountry_IgnoresInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	active := f.makeEmployee("Active", "Person", 100000)
	f.mustCreateEmployee(t, active)
	inactive := f.makeEmployee("Inactive", "Person", 999999)
	f.mustCreateEmployee(t, inactive)
	if err := f.repo.Delete(ctx, inactive.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	sr, _ := f.repo.GetSalaryRangeByCountry(ctx, "United States")
	if sr.Count != 1 || sr.Average != 100000 {
		t.Errorf("inactive salary should be excluded; got count=%d avg=%f", sr.Count, sr.Average)
	}
}

func TestEmployee_GetSalaryRangeByCountry_UnknownCountry_ReturnsZeros(t *testing.T) {
	f := newEmployeeFixture(t)
	sr, err := f.repo.GetSalaryRangeByCountry(context.Background(), "Atlantis")
	if err != nil {
		t.Fatalf("GetSalaryRangeByCountry: %v", err)
	}
	if sr.Count != 0 || sr.Min != 0 || sr.Max != 0 || sr.Average != 0 || sr.Median != 0 {
		t.Errorf("expected all-zeros for unknown country, got %+v", sr)
	}
	if sr.Country != "Atlantis" {
		t.Errorf("Country echoed = %q, want Atlantis", sr.Country)
	}
}

func TestEmployee_GetSalaryByTitle_ComputesAverages(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i, s := range []float64{100000, 120000} {
		e := f.makeEmployee(fmt.Sprintf("Eng%d", i), "Person", s)
		e.Email = fmt.Sprintf("eng_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}
	mkt := f.makeEmployee("Mkt", "Person", 150000)
	mkt.Email = "mkt@example.com"
	mkt.JobTitleID = f.mktMgrTitleID
	f.mustCreateEmployee(t, mkt)

	sbt, err := f.repo.GetSalaryByTitle(ctx, "United States", "Software Engineer")
	if err != nil {
		t.Fatalf("GetSalaryByTitle: %v", err)
	}
	if sbt.Average != 110000 {
		t.Errorf("Average = %f, want 110000", sbt.Average)
	}
	if sbt.Count != 2 || sbt.Min != 100000 || sbt.Max != 120000 {
		t.Errorf("got %+v, want count=2 min=100000 max=120000", sbt)
	}
}

func TestEmployee_GetSalaryByTitle_IgnoresInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	keep := f.makeEmployee("Active", "Eng", 100000)
	f.mustCreateEmployee(t, keep)
	gone := f.makeEmployee("Gone", "Eng", 200000)
	f.mustCreateEmployee(t, gone)
	if err := f.repo.Delete(ctx, gone.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	sbt, _ := f.repo.GetSalaryByTitle(ctx, "United States", "Software Engineer")
	if sbt.Count != 1 || sbt.Average != 100000 {
		t.Errorf("inactive should be excluded; got count=%d avg=%f", sbt.Count, sbt.Average)
	}
}

func TestEmployee_GetSalaryByTitle_NoMatch_ReturnsZeros(t *testing.T) {
	f := newEmployeeFixture(t)
	sbt, err := f.repo.GetSalaryByTitle(context.Background(), "Atlantis", "Wizard")
	if err != nil {
		t.Fatalf("GetSalaryByTitle: %v", err)
	}
	if sbt.Count != 0 || sbt.Average != 0 {
		t.Errorf("expected zeros for no match, got %+v", sbt)
	}
}

func TestEmployee_GetDepartmentStats_OrderedByAvgDesc(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	// Engineering: 100k, 120k -> avg 110k
	for i, s := range []float64{100000, 120000} {
		e := f.makeEmployee(fmt.Sprintf("E%d", i), "Eng", s)
		e.Email = fmt.Sprintf("eng_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}
	// Marketing: 80k -> avg 80k
	m := f.makeEmployee("M", "Mkt", 80000)
	m.Email = "m@example.com"
	m.JobTitleID = f.mktMgrTitleID
	f.mustCreateEmployee(t, m)

	stats, err := f.repo.GetDepartmentStats(ctx, "United States")
	if err != nil {
		t.Fatalf("GetDepartmentStats: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("got %d departments, want 2", len(stats))
	}
	if stats[0].Department != "Engineering" {
		t.Errorf("first = %q, want Engineering (highest avg)", stats[0].Department)
	}
	if stats[0].EmployeeCount != 2 {
		t.Errorf("Engineering count = %d, want 2", stats[0].EmployeeCount)
	}
	if stats[1].Department != "Marketing" {
		t.Errorf("second = %q, want Marketing", stats[1].Department)
	}
}

func TestEmployee_GetDepartmentStats_NoCountryFilter_AggregatesGlobally(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	usa := f.makeEmployee("USA", "Eng", 100000)
	f.mustCreateEmployee(t, usa)
	india := f.makeEmployee("India", "Eng", 80000)
	india.Email = "india@example.com"
	india.CountryID = f.indiaID
	f.mustCreateEmployee(t, india)

	stats, err := f.repo.GetDepartmentStats(ctx, "")
	if err != nil {
		t.Fatalf("GetDepartmentStats: %v", err)
	}
	if len(stats) != 1 || stats[0].Department != "Engineering" {
		t.Fatalf("expected single Engineering row, got %v", stats)
	}
	if stats[0].EmployeeCount != 2 {
		t.Errorf("global EmployeeCount = %d, want 2 (both countries)", stats[0].EmployeeCount)
	}
}

func TestEmployee_GetDepartmentStats_IgnoresInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	a := f.makeEmployee("Active", "Eng", 100000)
	f.mustCreateEmployee(t, a)
	b := f.makeEmployee("Gone", "Eng", 999999)
	f.mustCreateEmployee(t, b)
	if err := f.repo.Delete(ctx, b.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	stats, _ := f.repo.GetDepartmentStats(ctx, "United States")
	if stats[0].EmployeeCount != 1 || stats[0].AverageSalary != 100000 {
		t.Errorf("inactive should be excluded; got count=%d avg=%f",
			stats[0].EmployeeCount, stats[0].AverageSalary)
	}
}

func TestEmployee_GetOrgSummary_AggregatesAcrossCountries(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		e := f.makeEmployee(fmt.Sprintf("USA%d", i), "Person", 100000)
		e.Email = fmt.Sprintf("usa_%d@example.com", i)
		f.mustCreateEmployee(t, e)
	}
	india := f.makeEmployee("Indian", "Person", 100000)
	india.Email = "india@example.com"
	india.CountryID = f.indiaID
	f.mustCreateEmployee(t, india)

	summary, err := f.repo.GetOrgSummary(ctx)
	if err != nil {
		t.Fatalf("GetOrgSummary: %v", err)
	}
	if summary.TotalEmployees != 3 {
		t.Errorf("TotalEmployees = %d, want 3", summary.TotalEmployees)
	}
	if summary.AverageSalary != 100000 {
		t.Errorf("AverageSalary = %f, want 100000", summary.AverageSalary)
	}
	if summary.TotalCountries != 2 {
		t.Errorf("TotalCountries = %d, want 2", summary.TotalCountries)
	}
	if summary.TotalDepartments != 1 {
		t.Errorf("TotalDepartments = %d, want 1", summary.TotalDepartments)
	}
	if len(summary.CountryBreakdown) != 2 {
		t.Errorf("CountryBreakdown has %d entries, want 2", len(summary.CountryBreakdown))
	}
	// Ordered by COUNT(*) DESC -> USA (2) then India (1)
	if summary.CountryBreakdown[0].EmployeeCount < summary.CountryBreakdown[1].EmployeeCount {
		t.Errorf("CountryBreakdown should be ordered DESC by count: %v", summary.CountryBreakdown)
	}
}

func TestEmployee_GetOrgSummary_EmptyDB_ReturnsZeros(t *testing.T) {
	f := newEmployeeFixture(t)
	summary, err := f.repo.GetOrgSummary(context.Background())
	if err != nil {
		t.Fatalf("GetOrgSummary: %v", err)
	}
	if summary.TotalEmployees != 0 || summary.AverageSalary != 0 {
		t.Errorf("expected zero totals on empty DB, got %+v", summary)
	}
	if len(summary.CountryBreakdown) != 0 {
		t.Errorf("expected empty CountryBreakdown, got %v", summary.CountryBreakdown)
	}
}

func TestEmployee_GetOrgSummary_IgnoresInactive(t *testing.T) {
	f := newEmployeeFixture(t)
	ctx := context.Background()
	keep := f.makeEmployee("Active", "Person", 100000)
	f.mustCreateEmployee(t, keep)
	gone := f.makeEmployee("Inactive", "Person", 200000)
	f.mustCreateEmployee(t, gone)
	if err := f.repo.Delete(ctx, gone.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	summary, _ := f.repo.GetOrgSummary(ctx)
	if summary.TotalEmployees != 1 || summary.AverageSalary != 100000 {
		t.Errorf("inactive should be excluded; got total=%d avg=%f",
			summary.TotalEmployees, summary.AverageSalary)
	}
}
