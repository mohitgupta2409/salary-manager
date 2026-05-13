package model

import "time"

type Employee struct {
	ID         int64     `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	JobTitle   string    `json:"job_title"`
	Department string    `json:"department"`
	Country    string    `json:"country"`
	Salary     float64   `json:"salary"`
	Currency   string    `json:"currency"`
	JoinDate   time.Time `json:"join_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type EmployeeFilter struct {
	Search   string `json:"search"`
	Country  string `json:"country"`
	JobTitle string `json:"job_title"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

type EmployeeListResult struct {
	Employees []Employee `json:"employees"`
	Total     int64      `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

type SalaryRange struct {
	Country  string  `json:"country"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Average  float64 `json:"average"`
	Median   float64 `json:"median"`
	Count    int64   `json:"count"`
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
	TotalEmployees    int64              `json:"total_employees"`
	AverageSalary     float64            `json:"average_salary"`
	TotalCountries    int64              `json:"total_countries"`
	TotalDepartments  int64              `json:"total_departments"`
	CountryBreakdown  []CountryHeadcount `json:"country_breakdown"`
}
