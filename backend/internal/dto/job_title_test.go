package dto

import (
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

func TestToJobTitleResponse(t *testing.T) {
	got := ToJobTitleResponse(&model.JobTitle{
		ID: 5, Name: "DevOps", DepartmentID: 2,
		Department: "Platform", IsActive: true,
	})
	if got == nil {
		t.Fatal("got nil")
	}
	if got.ID != 5 || got.Name != "DevOps" || got.DepartmentID != 2 || got.Department != "Platform" {
		t.Errorf("conversion wrong: %+v", got)
	}
}

func TestToJobTitleResponse_Nil(t *testing.T) {
	if got := ToJobTitleResponse(nil); got != nil {
		t.Errorf("nil input should yield nil, got %+v", got)
	}
}

func TestToJobTitleListResponse(t *testing.T) {
	in := &model.JobTitleListResult{
		JobTitles: []model.JobTitle{
			{ID: 1, Name: "Eng", DepartmentID: 1, Department: "Engineering"},
		},
		Total: 1, Limit: 50, Offset: 0,
	}
	got := ToJobTitleListResponse(in)
	if len(got.JobTitles) != 1 || got.JobTitles[0].Department != "Engineering" {
		t.Errorf("conversion wrong: %+v", got)
	}
}

func TestToJobTitleListResponse_Nil(t *testing.T) {
	got := ToJobTitleListResponse(nil)
	if got == nil || got.JobTitles == nil {
		t.Error("expected non-nil response with empty slice")
	}
}

func TestToModelJobTitle(t *testing.T) {
	got := ToModelJobTitle(&JobTitleCreateRequest{Name: "QA", DepartmentID: 3})
	if got.Name != "QA" || got.DepartmentID != 3 {
		t.Errorf("conversion wrong: %+v", got)
	}
}

func TestToModelJobTitle_Nil(t *testing.T) {
	if got := ToModelJobTitle(nil); got == nil {
		t.Error("expected zero-value model, got nil")
	}
}
