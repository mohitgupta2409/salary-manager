package dto

import (
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// JobTitleCreateRequest is the body of POST /api/job-titles.
type JobTitleCreateRequest struct {
	Name         string `json:"name"`
	DepartmentID int64  `json:"department_id"`
}

// JobTitleResponse is what the API returns. The Department string carries
// the human-readable department name (denormalized via JOIN at the
// repository layer); DepartmentID is also returned so clients can issue
// follow-up queries without a second lookup.
type JobTitleResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	DepartmentID int64     `json:"department_id"`
	Department   string    `json:"department,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// JobTitleListResponse wraps a page of job titles with pagination metadata.
type JobTitleListResponse struct {
	JobTitles []JobTitleResponse `json:"job_titles"`
	Total     int64              `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

func ToJobTitleResponse(jt *model.JobTitle) *JobTitleResponse {
	if jt == nil {
		return nil
	}
	return &JobTitleResponse{
		ID:           jt.ID,
		Name:         jt.Name,
		DepartmentID: jt.DepartmentID,
		Department:   jt.Department,
		IsActive:     jt.IsActive,
		CreatedAt:    jt.CreatedAt,
		UpdatedAt:    jt.UpdatedAt,
	}
}

func ToJobTitleListResponse(r *model.JobTitleListResult) *JobTitleListResponse {
	if r == nil {
		return &JobTitleListResponse{JobTitles: []JobTitleResponse{}}
	}
	out := &JobTitleListResponse{
		JobTitles: make([]JobTitleResponse, 0, len(r.JobTitles)),
		Total:     r.Total,
		Limit:     r.Limit,
		Offset:    r.Offset,
	}
	for i := range r.JobTitles {
		out.JobTitles = append(out.JobTitles, *ToJobTitleResponse(&r.JobTitles[i]))
	}
	return out
}

func ToModelJobTitle(req *JobTitleCreateRequest) *model.JobTitle {
	if req == nil {
		return &model.JobTitle{}
	}
	return &model.JobTitle{
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
	}
}
