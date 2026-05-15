package sqlite

// This file contains shared test utilities used by every *_test.go file in
// this package: the in-memory test database constructor, helpers for
// flipping is_active flags on rows that have no production "delete" path,
// the employeeFixture (which seeds a small reference catalog), and small
// model factory helpers used to build valid struct values.
//
// Putting these here keeps each per-entity *_test.go file focused purely on
// scenarios for that entity.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// =============================================================================
// In-memory database setup
// =============================================================================

// newTestDB opens an in-memory SQLite database and applies the migration
// schema owned by the seed/migration tool (cmd/seed/schema.sql).
//
// InitDB itself does not create any schema; it only opens a connection. So
// repository tests use this helper to provision an isolated, fully-migrated
// database without coupling the production runtime code to migrations.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}

	schema, err := os.ReadFile(schemaPath(t))
	if err != nil {
		db.Close()
		t.Fatalf("read schema: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		t.Fatalf("apply schema: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// schemaPath resolves the canonical schema.sql relative to this file so the
// helper works regardless of where `go test` is invoked from.
func schemaPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file location")
	}
	// internal/repository/sqlite/testutil_test.go -> backend/cmd/seed/schema.sql
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "cmd", "seed", "schema.sql")
}

// =============================================================================
// Direct-SQL helpers for marking rows inactive
//
// The production reference repositories (Country, Department, JobTitle) have
// no Delete/Deactivate method, but list-method tests that exercise the
// IncludeInactive filter need a way to flip the is_active flag.
// =============================================================================

func deactivateCountry(t *testing.T, db *sql.DB, id int64) {
	t.Helper()
	if _, err := db.Exec(`UPDATE countries SET is_active = 0 WHERE id = ?`, id); err != nil {
		t.Fatalf("deactivate country %d: %v", id, err)
	}
}

func deactivateDepartment(t *testing.T, db *sql.DB, id int64) {
	t.Helper()
	if _, err := db.Exec(`UPDATE departments SET is_active = 0 WHERE id = ?`, id); err != nil {
		t.Fatalf("deactivate department %d: %v", id, err)
	}
}

func deactivateJobTitle(t *testing.T, db *sql.DB, id int64) {
	t.Helper()
	if _, err := db.Exec(`UPDATE job_titles SET is_active = 0 WHERE id = ?`, id); err != nil {
		t.Fatalf("deactivate job title %d: %v", id, err)
	}
}

// =============================================================================
// Model factory helpers
//
// These return ready-to-insert struct values with sensible defaults so tests
// don't have to spell every field. Tests can override fields after the fact.
// =============================================================================

// newCountry returns a Country with active defaults. Code/Currency are passed
// in lowercase intentionally so tests can verify normalization paths.
func newCountry(name, code, currency string) *model.Country {
	return &model.Country{Name: name, Code: code, Currency: currency}
}

// newDepartment returns a Department with the given name.
func newDepartment(name string) *model.Department {
	return &model.Department{Name: name}
}

// newJobTitle returns a JobTitle attached to the given department id.
func newJobTitle(name string, departmentID int64) *model.JobTitle {
	return &model.JobTitle{Name: name, DepartmentID: departmentID}
}

// =============================================================================
// employeeFixture — pre-seeded reference catalog for employee tests
// =============================================================================

// employeeFixture is a fully-seeded reference catalog plus a fresh employee
// repo, used by employee_test.go. Two countries, two departments, and two
// job titles are pre-seeded so tests can exercise filtering and FK paths.
type employeeFixture struct {
	db            *sql.DB
	repo          *employeeRepo
	usaID         int64
	indiaID       int64
	engID         int64
	mktID         int64
	swEngTitleID  int64
	mktMgrTitleID int64
}

func newEmployeeFixture(t *testing.T) *employeeFixture {
	t.Helper()
	db := newTestDB(t)
	ctx := context.Background()
	cr := &countryRepo{db: db}
	dr := &departmentRepo{db: db}
	jr := &jobTitleRepo{db: db}

	usa := newCountry("United States", "US", "USD")
	if err := cr.Create(ctx, usa); err != nil {
		t.Fatalf("seed USA: %v", err)
	}
	india := newCountry("India", "IN", "INR")
	if err := cr.Create(ctx, india); err != nil {
		t.Fatalf("seed India: %v", err)
	}

	eng := newDepartment("Engineering")
	if err := dr.Create(ctx, eng); err != nil {
		t.Fatalf("seed Engineering: %v", err)
	}
	mkt := newDepartment("Marketing")
	if err := dr.Create(ctx, mkt); err != nil {
		t.Fatalf("seed Marketing: %v", err)
	}

	swEng := newJobTitle("Software Engineer", eng.ID)
	if err := jr.Create(ctx, swEng); err != nil {
		t.Fatalf("seed Software Engineer: %v", err)
	}
	mktMgr := newJobTitle("Marketing Manager", mkt.ID)
	if err := jr.Create(ctx, mktMgr); err != nil {
		t.Fatalf("seed Marketing Manager: %v", err)
	}

	return &employeeFixture{
		db:            db,
		repo:          &employeeRepo{db: db},
		usaID:         usa.ID,
		indiaID:       india.ID,
		engID:         eng.ID,
		mktID:         mkt.ID,
		swEngTitleID:  swEng.ID,
		mktMgrTitleID: mktMgr.ID,
	}
}

// makeEmployee returns a fully-populated Employee that is valid against the
// fixture's seeded reference data. Tests can mutate the result before
// calling repo.Create.
func (f *employeeFixture) makeEmployee(first, last string, salary float64) *model.Employee {
	return &model.Employee{
		FirstName:  first,
		LastName:   last,
		Email:      fmt.Sprintf("%s.%s@example.com", first, last),
		JobTitleID: f.swEngTitleID,
		CountryID:  f.usaID,
		Salary:     salary,
		Address:    "1 Test Way",
		JoinDate:   time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

// mustCreateEmployee inserts the employee and fails the test on error.
func (f *employeeFixture) mustCreateEmployee(t *testing.T, e *model.Employee) {
	t.Helper()
	if err := f.repo.Create(context.Background(), e); err != nil {
		t.Fatalf("create employee %s: %v", e.Email, err)
	}
}
