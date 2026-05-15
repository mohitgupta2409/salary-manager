package dto

import (
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// EmployeeCreateRequest is the body of POST /api/employees. Foreign keys
// are referenced by id; the API enriches them on the response side.
type EmployeeCreateRequest struct {
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	JobTitleID int64     `json:"job_title_id"`
	CountryID  int64     `json:"country_id"`
	Salary     float64   `json:"salary"`
	Address    string    `json:"address,omitempty"`
	JoinDate   time.Time `json:"join_date"`
}

// EmployeeUpdateRequest is the body of PUT /api/employees/{id}. The {id}
// path parameter, not this struct, identifies the row being updated.
type EmployeeUpdateRequest = EmployeeCreateRequest

// EmployeeResponse is the API view of an employee. It is intentionally
// trimmed compared with the DB-layer model.Employee to expose only the
// fields a client actually needs to render a row:
//
//   - FullName replaces FirstName/LastName
//   - Country (name) replaces CountryID; Currency (ISO-4217) is included
//     so clients can render the salary in the employee's local currency
//     without an extra lookup
//   - JobTitle (name) replaces JobTitleID
//   - Department (name) is kept for context
//
// Database-only metadata (FK ids) lives on model.Employee and is not
// surfaced through the API.
type EmployeeResponse struct {
	ID         int64     `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Salary     float64   `json:"salary"`
	Currency   string    `json:"currency,omitempty"`
	Address    string    `json:"address,omitempty"`
	JoinDate   time.Time `json:"join_date"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Country    string    `json:"country,omitempty"`
	JobTitle   string    `json:"job_title,omitempty"`
	Department string    `json:"department,omitempty"`
}

// EmployeeListResponse wraps a page of employees with pagination metadata.
type EmployeeListResponse struct {
	Employees []EmployeeResponse `json:"employees"`
	Total     int64              `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

func ToEmployeeResponse(e *model.Employee) *EmployeeResponse {
	if e == nil {
		return nil
	}
	return &EmployeeResponse{
		ID:         e.ID,
		FullName:   e.FullName(),
		Email:      e.Email,
		Salary:     e.Salary,
		Currency:   e.Currency,
		Address:    e.Address,
		JoinDate:   e.JoinDate,
		IsActive:   e.IsActive,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
		Country:    e.Country,
		JobTitle:   e.JobTitle,
		Department: e.Department,
	}
}

func ToEmployeeListResponse(r *model.EmployeeListResult) *EmployeeListResponse {
	if r == nil {
		return &EmployeeListResponse{Employees: []EmployeeResponse{}}
	}
	out := &EmployeeListResponse{
		Employees: make([]EmployeeResponse, 0, len(r.Employees)),
		Total:     r.Total,
		Limit:     r.Limit,
		Offset:    r.Offset,
	}
	for i := range r.Employees {
		out.Employees = append(out.Employees, *ToEmployeeResponse(&r.Employees[i]))
	}
	return out
}

// ToModelEmployee builds a writable DB-layer Employee value from a request.
// Validation, normalisation, and FK lookup live in the service layer.
func ToModelEmployee(req *EmployeeCreateRequest) *model.Employee {
	if req == nil {
		return &model.Employee{}
	}
	return &model.Employee{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		JobTitleID: req.JobTitleID,
		CountryID:  req.CountryID,
		Salary:     req.Salary,
		Address:    req.Address,
		JoinDate:   req.JoinDate,
	}
}
