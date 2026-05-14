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
