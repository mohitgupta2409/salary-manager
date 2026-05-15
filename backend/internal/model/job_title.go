package model

import "time"

type JobTitle struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	DepartmentID int64     `json:"department_id"`
	Department   string    `json:"department,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// JobTitleListRequest is the input to JobTitleRepository.List. DepartmentID
// is an optional filter — zero means "any department".
type JobTitleListRequest struct {
	Pagination
	DepartmentID    int64 `json:"department_id"`
	IncludeInactive bool  `json:"include_inactive"`
}

// JobTitleListResult is a page of job titles plus the unfiltered total.
type JobTitleListResult struct {
	JobTitles []JobTitle `json:"job_titles"`
	Total     int64      `json:"total"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}
