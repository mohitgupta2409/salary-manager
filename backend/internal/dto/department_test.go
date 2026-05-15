package dto

import (
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

func TestToDepartmentResponse(t *testing.T) {
	got := ToDepartmentResponse(&model.Department{ID: 3, Name: "Sales", IsActive: true})
	if got == nil {
		t.Fatal("got nil")
	}
	if got.ID != 3 || got.Name != "Sales" || !got.IsActive {
		t.Errorf("conversion wrong: %+v", got)
	}
}

func TestToDepartmentResponse_Nil(t *testing.T) {
	if got := ToDepartmentResponse(nil); got != nil {
		t.Errorf("nil input should yield nil, got %+v", got)
	}
}

func TestToDepartmentListResponse(t *testing.T) {
	in := &model.DepartmentListResult{
		Departments: []model.Department{{ID: 1, Name: "Eng"}, {ID: 2, Name: "HR"}},
		Total:       2, Limit: 5, Offset: 0,
	}
	got := ToDepartmentListResponse(in)
	if len(got.Departments) != 2 || got.Total != 2 || got.Limit != 5 {
		t.Errorf("conversion wrong: %+v", got)
	}
}

func TestToDepartmentListResponse_Nil(t *testing.T) {
	got := ToDepartmentListResponse(nil)
	if got == nil || got.Departments == nil {
		t.Error("expected non-nil response with empty slice")
	}
}

func TestToModelDepartment(t *testing.T) {
	got := ToModelDepartment(&DepartmentCreateRequest{Name: "Marketing"})
	if got.Name != "Marketing" {
		t.Errorf("Name = %q", got.Name)
	}
}

func TestToModelDepartment_Nil(t *testing.T) {
	if got := ToModelDepartment(nil); got == nil {
		t.Error("expected zero-value model, got nil")
	}
}
