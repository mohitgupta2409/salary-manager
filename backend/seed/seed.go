package seed

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

// Repos bundles the repositories the seeder needs.
type Repos struct {
	Country   repository.CountryRepository
	Dept      repository.DepartmentRepository
	JobTitle  repository.JobTitleRepository
	Employee  repository.EmployeeRepository
}

// SeedReferenceData populates countries, departments, and job_titles.
// Returns lookup maps that can be used by SeedEmployees to pick valid FKs.
type referenceLookup struct {
	countriesByName map[string]int64    // country name -> id
	jobTitleIDs     []int64             // pool of job_title ids
	jobTitleByName  map[string]int64    // "Department/Title" -> id (for tests)
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

// SeedEmployees creates `count` employees referencing the seeded reference
// data. Uses a deterministic seed so output is reproducible.
func SeedEmployees(ctx context.Context, r Repos, lookup *referenceLookup, count int) error {
	rng := rand.New(rand.NewSource(42))

	totalWeight := 0
	for _, c := range Countries {
		totalWeight += c.Weight
	}

	emailSet := make(map[string]bool)

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

		emp := &model.Employee{
			FirstName:  first,
			LastName:   last,
			Email:      email,
			JobTitleID: jobTitleID,
			CountryID:  countryID,
			Salary:     salary,
			Address:    address,
			JoinDate:   time.Date(joinYear, joinMonth, joinDay, 0, 0, 0, 0, time.UTC),
		}

		if err := r.Employee.Create(ctx, emp); err != nil {
			return fmt.Errorf("create employee %d: %w", i, err)
		}

		if (i+1)%1000 == 0 {
			fmt.Printf("  Seeded %d/%d employees\n", i+1, count)
		}
	}
	return nil
}

// SeedAll seeds reference data and the requested number of employees.
func SeedAll(ctx context.Context, r Repos, employeeCount int) error {
	lookup, err := SeedReferenceData(ctx, r)
	if err != nil {
		return err
	}
	return SeedEmployees(ctx, r, lookup, employeeCount)
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
