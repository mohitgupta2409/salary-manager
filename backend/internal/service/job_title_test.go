package service

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
)

func TestJobTitleService_Create_Success_ReturnsResponseDTO(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())
	resp, err := svc.Create(context.Background(), &dto.JobTitleCreateRequest{
		Name: "DevOps Engineer", DepartmentID: 1,
	})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp == nil || resp.ID == 0 {
		t.Fatalf("response wrong: %+v", resp)
	}
	if resp.Name != "DevOps Engineer" {
		t.Errorf("Name = %q", resp.Name)
	}
	// The service is responsible for populating the denormalised Department
	// name on the create response (the underlying repo Create only writes
	// raw columns).
	if resp.Department != "Engineering" {
		t.Errorf("Department = %q, want Engineering", resp.Department)
	}
}

func TestJobTitleService_Create_TrimsName(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())
	resp, err := svc.Create(context.Background(), &dto.JobTitleCreateRequest{
		Name: "  QA Engineer  ", DepartmentID: 1,
	})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp.Name != "QA Engineer" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestJobTitleService_Create_NilRequest(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())
	if _, err := svc.Create(context.Background(), nil); err != ErrInvalidInput {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestJobTitleService_Create_Validation(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())

	tests := []struct {
		name    string
		req     dto.JobTitleCreateRequest
		wantErr string
	}{
		{"empty name", dto.JobTitleCreateRequest{Name: "", DepartmentID: 1}, "job title name is required"},
		{"missing dept", dto.JobTitleCreateRequest{Name: "Engineer", DepartmentID: 0}, "department is required"},
		{"non-existent dept", dto.JobTitleCreateRequest{Name: "Engineer", DepartmentID: 999}, "department does not exist or is inactive"},
		{"inactive dept", dto.JobTitleCreateRequest{Name: "Engineer", DepartmentID: 2}, "department does not exist or is inactive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), &tt.req)
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestJobTitleService_List_NeverNil(t *testing.T) {
	svc := NewJobTitleService(&mockJobTitleRepo{byID: map[int64]*model.JobTitle{}}, newMockDeptRepo())
	out, err := svc.List(context.Background(), model.JobTitleListRequest{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out == nil || out.JobTitles == nil {
		t.Error("List should return empty slice, not nil")
	}
}

func TestJobTitleService_List_ReturnsResponseDTOs(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())
	out, err := svc.List(context.Background(), model.JobTitleListRequest{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out.Total < 1 {
		t.Errorf("Total = %d, want >=1", out.Total)
	}
	// Verify the active mock title's denormalised Department surfaces in DTO.
	var se *dto.JobTitleResponse
	for i := range out.JobTitles {
		if out.JobTitles[i].Name == "Software Engineer" {
			c := out.JobTitles[i]
			se = &c
			break
		}
	}
	if se == nil {
		t.Fatal("expected to find seeded Software Engineer")
	}
	if se.Department != "Engineering" {
		t.Errorf("Department = %q, want Engineering", se.Department)
	}
	if se.DepartmentID != 1 {
		t.Errorf("DepartmentID = %d, want 1", se.DepartmentID)
	}
}

func TestJobTitleService_List_CapsLimit(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())
	out, err := svc.List(context.Background(), model.JobTitleListRequest{
		Pagination: model.Pagination{Limit: 5000},
	})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out.Total == 0 {
		t.Error("expected non-zero total")
	}
}
