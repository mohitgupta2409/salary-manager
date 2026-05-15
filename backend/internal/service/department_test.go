package service

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
)

func TestDepartmentService_Create_Success_ReturnsResponseDTO(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	resp, err := svc.Create(context.Background(), &dto.DepartmentCreateRequest{Name: "Sales"})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp == nil || resp.ID == 0 || resp.Name != "Sales" {
		t.Errorf("response wrong: %+v", resp)
	}
}

func TestDepartmentService_Create_TrimsName(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	resp, err := svc.Create(context.Background(), &dto.DepartmentCreateRequest{Name: "  Marketing  "})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp.Name != "Marketing" {
		t.Errorf("Name = %q, want Marketing", resp.Name)
	}
}

func TestDepartmentService_Create_NilRequest(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	if _, err := svc.Create(context.Background(), nil); err != ErrInvalidInput {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestDepartmentService_Create_Validation(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	for _, name := range []string{"", "   "} {
		_, err := svc.Create(context.Background(), &dto.DepartmentCreateRequest{Name: name})
		if err == nil || err.Error() != "department name is required" {
			t.Errorf("for %q, error = %v, want 'department name is required'", name, err)
		}
	}
}

func TestDepartmentService_List_NeverNil(t *testing.T) {
	svc := NewDepartmentService(&mockDeptRepo{byID: map[int64]*model.Department{}})
	out, err := svc.List(context.Background(), model.DepartmentListRequest{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out == nil || out.Departments == nil {
		t.Error("List should return empty slice, not nil")
	}
}

func TestDepartmentService_List_ReturnsResponseDTOs(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	out, err := svc.List(context.Background(), model.DepartmentListRequest{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out.Total < 1 {
		t.Errorf("Total = %d, want >=1", out.Total)
	}
	for _, d := range out.Departments {
		if d.ID == 0 || d.Name == "" {
			t.Errorf("DTO incomplete: %+v", d)
		}
	}
}

func TestDepartmentService_List_CapsLimit(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	out, err := svc.List(context.Background(), model.DepartmentListRequest{
		Pagination: model.Pagination{Limit: 9999},
	})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out.Total == 0 {
		t.Error("expected non-zero total")
	}
}
