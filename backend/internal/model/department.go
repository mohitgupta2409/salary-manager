package model

import "time"

type Department struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DepartmentListRequest is the input to DepartmentRepository.List.
type DepartmentListRequest struct {
	Pagination
	IncludeInactive bool `json:"include_inactive"`
}

// DepartmentListResult is a page of departments plus the unfiltered total.
type DepartmentListResult struct {
	Departments []Department `json:"departments"`
	Total       int64        `json:"total"`
	Limit       int          `json:"limit"`
	Offset      int          `json:"offset"`
}
