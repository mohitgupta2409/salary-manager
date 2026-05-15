package dto

import (
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// DepartmentCreateRequest is the body of POST /api/departments.
type DepartmentCreateRequest struct {
	Name string `json:"name"`
}

// DepartmentResponse is what the API returns for a single department.
type DepartmentResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DepartmentListResponse wraps a page of departments with pagination
// metadata.
type DepartmentListResponse struct {
	Departments []DepartmentResponse `json:"departments"`
	Total       int64                `json:"total"`
	Limit       int                  `json:"limit"`
	Offset      int                  `json:"offset"`
}

func ToDepartmentResponse(d *model.Department) *DepartmentResponse {
	if d == nil {
		return nil
	}
	return &DepartmentResponse{
		ID:        d.ID,
		Name:      d.Name,
		IsActive:  d.IsActive,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func ToDepartmentListResponse(r *model.DepartmentListResult) *DepartmentListResponse {
	if r == nil {
		return &DepartmentListResponse{Departments: []DepartmentResponse{}}
	}
	out := &DepartmentListResponse{
		Departments: make([]DepartmentResponse, 0, len(r.Departments)),
		Total:       r.Total,
		Limit:       r.Limit,
		Offset:      r.Offset,
	}
	for i := range r.Departments {
		out.Departments = append(out.Departments, *ToDepartmentResponse(&r.Departments[i]))
	}
	return out
}

func ToModelDepartment(req *DepartmentCreateRequest) *model.Department {
	if req == nil {
		return &model.Department{}
	}
	return &model.Department{Name: req.Name}
}
