package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

// Repos bundles the repositories the seeder needs.
type Repos struct {
	Country  repository.CountryRepository
	Dept     repository.DepartmentRepository
	JobTitle repository.JobTitleRepository
	Employee repository.EmployeeRepository
}

// SeedReferenceData populates countries, departments, and job_titles.
// Returns lookup maps that can be used by SeedEmployees to pick valid FKs.
type referenceLookup struct {
	countriesByName map[string]int64 // country name -> id
	jobTitleIDs     []int64          // pool of job_title ids
	jobTitleByName  map[string]int64 // "Department/Title" -> id (for tests)
}

func SeedReferenceData(ctx context.Context, r Repos) (*referenceLookup, error) {
	lookup := &referenceLookup{
		countriesByName: make(map[string]int64),
		jobTitleByName:  make(map[string]int64),
	}

	for _, c := range Countries {
		country := &model.Country{
			Name: c.Name, Code: c.Code, Currency: c.Currency,
		}
		if err := r.Country.Create(ctx, country); err != nil {
			return nil, fmt.Errorf("seed country %q: %w", c.Name, err)
		}
		lookup.countriesByName[c.Name] = country.ID
	}

	deptIDByName := make(map[string]int64)
	for _, name := range Departments {
		d := &model.Department{Name: name}
		if err := r.Dept.Create(ctx, d); err != nil {
			return nil, fmt.Errorf("seed department %q: %w", name, err)
		}
		deptIDByName[name] = d.ID
	}

	for deptName, titles := range JobTitlesByDepartment {
		deptID, ok := deptIDByName[deptName]
		if !ok {
			return nil, fmt.Errorf("department %q referenced by job titles but not seeded", deptName)
		}
		for _, title := range titles {
			jt := &model.JobTitle{Name: title, DepartmentID: deptID}
			if err := r.JobTitle.Create(ctx, jt); err != nil {
				return nil, fmt.Errorf("seed job title %q: %w", title, err)
			}
			lookup.jobTitleIDs = append(lookup.jobTitleIDs, jt.ID)
			lookup.jobTitleByName[deptName+"/"+title] = jt.ID
		}
	}

	return lookup, nil
}

// Bulk insert tuning. colsPerRow must match the column list in
// buildEmployeeInsertSQL (is_active is hard-coded in the literal, not bound).
// rowsPerBatch * colsPerRow must stay well below SQLite's
// SQLITE_MAX_VARIABLE_NUMBER (32,766 on modern builds).
const (
	colsPerRow   = 8
	rowsPerBatch = 500
)

// buildEmployeeInsertSQL returns an INSERT statement with `rows` VALUES tuples.
// We omit created_at/updated_at so SQLite uses the DEFAULT CURRENT_TIMESTAMP
// from the schema, saving two bound params per row and a time.Now call.
func buildEmployeeInsertSQL(rows int) string {
	const tuple = "(?, ?, ?, ?, ?, ?, ?, ?, 1)"
	values := make([]string, rows)
	for i := range values {
		values[i] = tuple
	}
	return `INSERT INTO employees
		(first_name, last_name, email, job_title_id, country_id,
		 salary, address, join_date, is_active)
		VALUES ` + strings.Join(values, ",")
}

// SeedEmployees creates `count` employees referencing the seeded reference
// data, in a single transaction using batched multi-row INSERTs. Uses a
// deterministic RNG seed so output is reproducible.
func SeedEmployees(ctx context.Context, db *sql.DB, lookup *referenceLookup, count int) error {
	if count <= 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(42))

	totalWeight := 0
	for _, c := range Countries {
		totalWeight += c.Weight
	}

	emailSet := make(map[string]bool, count)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	fullStmt, err := tx.PrepareContext(ctx, buildEmployeeInsertSQL(rowsPerBatch))
	if err != nil {
		return fmt.Errorf("prepare full batch: %w", err)
	}
	defer fullStmt.Close()

	var tailStmt *sql.Stmt
	if tail := count % rowsPerBatch; tail > 0 {
		tailStmt, err = tx.PrepareContext(ctx, buildEmployeeInsertSQL(tail))
		if err != nil {
			return fmt.Errorf("prepare tail batch: %w", err)
		}
		defer tailStmt.Close()
	}

	args := make([]any, 0, rowsPerBatch*colsPerRow)
	rowsInBatch := 0

	flush := func(stmt *sql.Stmt, upTo int) error {
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return fmt.Errorf("insert batch ending at row %d: %w", upTo, err)
		}
		args = args[:0]
		rowsInBatch = 0
		return nil
	}

	for i := 0; i < count; i++ {
		first := FirstNames[rng.Intn(len(FirstNames))]
		last := LastNames[rng.Intn(len(LastNames))]
		email := uniqueEmail(first, last, i, emailSet)

		country := pickCountry(rng, totalWeight)
		countryID := lookup.countriesByName[country.Name]
		jobTitleID := lookup.jobTitleIDs[rng.Intn(len(lookup.jobTitleIDs))]

		salary := country.MinBase + rng.Float64()*(country.MaxBase-country.MinBase)
		salary = float64(int(salary/100)) * 100

		joinYear := 2015 + rng.Intn(10)
		joinMonth := time.Month(1 + rng.Intn(12))
		joinDay := 1 + rng.Intn(28)

		street := Streets[rng.Intn(len(Streets))]
		address := fmt.Sprintf("%d %s, %s", 100+rng.Intn(9000), street, country.Name)

		joinDate := time.Date(joinYear, joinMonth, joinDay, 0, 0, 0, 0, time.UTC)

		args = append(args,
			first, last, email,
			jobTitleID, countryID,
			salary, address, joinDate,
		)
		rowsInBatch++

		if rowsInBatch == rowsPerBatch {
			if err := flush(fullStmt, i+1); err != nil {
				return err
			}
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("  Seeded %d/%d employees\n", i+1, count)
		}
	}

	if rowsInBatch > 0 {
		if tailStmt == nil {
			return fmt.Errorf("internal error: %d rows left over but no tail statement prepared", rowsInBatch)
		}
		if err := flush(tailStmt, count); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit employees tx: %w", err)
	}
	return nil
}

// SeedAll seeds reference data and the requested number of employees.
func SeedAll(ctx context.Context, db *sql.DB, r Repos, employeeCount int) error {
	lookup, err := SeedReferenceData(ctx, r)
	if err != nil {
		return err
	}
	return SeedEmployees(ctx, db, lookup, employeeCount)
}

func uniqueEmail(first, last string, index int, seen map[string]bool) string {
	base := fmt.Sprintf("%s.%s", strings.ToLower(first), strings.ToLower(last))
	email := base + "@company.com"
	if !seen[email] {
		seen[email] = true
		return email
	}
	email = fmt.Sprintf("%s.%s%d@company.com", strings.ToLower(first), strings.ToLower(last), index)
	seen[email] = true
	return email
}

func pickCountry(rng *rand.Rand, totalWeight int) CountryConfig {
	r := rng.Intn(totalWeight)
	for _, c := range Countries {
		r -= c.Weight
		if r < 0 {
			return c
		}
	}
	return Countries[0]
}
