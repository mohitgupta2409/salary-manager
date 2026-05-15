package model

import "time"

type Employee struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	JobTitleID   int64     `json:"job_title_id"`
	CountryID    int64     `json:"country_id"`
	Salary       float64   `json:"salary"`
	Address      string    `json:"address,omitempty"`
	JoinDate     time.Time `json:"join_date"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Denormalized fields populated via JOINs (not persisted on Employee row)
	JobTitle   string `json:"job_title,omitempty"`
	Department string `json:"department,omitempty"`
	Country    string `json:"country,omitempty"`
	Currency   string `json:"currency,omitempty"`
}

func (e Employee) FullName() string {
	if e.FirstName == "" {
		return e.LastName
	}
	if e.LastName == "" {
		return e.FirstName
	}
	return e.FirstName + " " + e.LastName
}

// EmployeeFilter is the input to EmployeeRepository.List. It combines
// query filters with pagination (offset/limit, embedded). A zero or negative
// Limit means "return all matching records".
type EmployeeFilter struct {
	Search       string `json:"search"`
	CountryID    int64  `json:"country_id"`
	JobTitleID   int64  `json:"job_title_id"`
	DepartmentID int64  `json:"department_id"`
	Pagination
}

// EmployeeListResult is a page of employees plus the unfiltered total.
type EmployeeListResult struct {
	Employees []Employee `json:"employees"`
	Total     int64      `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type SalaryRange struct {
	Country string  `json:"country"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Average float64 `json:"average"`
	Median  float64 `json:"median"`
	Count   int64   `json:"count"`
}

type SalaryByTitle struct {
	Country  string  `json:"country"`
	JobTitle string  `json:"job_title"`
	Average  float64 `json:"average"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Count    int64   `json:"count"`
}

type DepartmentStats struct {
	Department    string  `json:"department"`
	AverageSalary float64 `json:"average_salary"`
	MinSalary     float64 `json:"min_salary"`
	MaxSalary     float64 `json:"max_salary"`
	EmployeeCount int64   `json:"employee_count"`
}

type CountryHeadcount struct {
	Country       string  `json:"country"`
	EmployeeCount int64   `json:"employee_count"`
	AverageSalary float64 `json:"average_salary"`
}

type OrgSummary struct {
	TotalEmployees   int64              `json:"total_employees"`
	AverageSalary    float64            `json:"average_salary"`
	TotalCountries   int64              `json:"total_countries"`
	TotalDepartments int64              `json:"total_departments"`
	CountryBreakdown []CountryHeadcount `json:"country_breakdown"`
}
