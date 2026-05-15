package service

import (
	"context"
	"errors"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
)

// =============================================================================
// Create
// =============================================================================

func TestEmployeeService_Create_Success_ReturnsResponseDTO(t *testing.T) {
	svc := newSvc()

	resp, err := svc.Create(context.Background(), validEmployeeRequest())
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp == nil {
		t.Fatal("Create returned nil response")
	}
	if resp.ID == 0 {
		t.Error("response should carry the assigned ID")
	}
	if resp.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want Jane Doe", resp.FullName)
	}
	if resp.Country != "United States" {
		t.Errorf("denormalised country missing: %+v", resp)
	}
	if resp.JobTitle != "Software Engineer" || resp.Department != "Engineering" {
		t.Errorf("denormalised job_title/department missing: %+v", resp)
	}
}

func TestEmployeeService_Create_NilRequest(t *testing.T) {
	svc := newSvc()
	_, err := svc.Create(context.Background(), nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestEmployeeService_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*dto.EmployeeCreateRequest)
		wantErr string
	}{
		{"empty first name", func(r *dto.EmployeeCreateRequest) { r.FirstName = "" }, "first name is required"},
		{"empty last name", func(r *dto.EmployeeCreateRequest) { r.LastName = "" }, "last name is required"},
		{"empty email", func(r *dto.EmployeeCreateRequest) { r.Email = "" }, "email is required"},
		{"invalid email", func(r *dto.EmployeeCreateRequest) { r.Email = "notanemail" }, "email must be valid"},
		{"missing job title id", func(r *dto.EmployeeCreateRequest) { r.JobTitleID = 0 }, "job title is required"},
		{"missing country id", func(r *dto.EmployeeCreateRequest) { r.CountryID = 0 }, "country is required"},
		{"negative salary", func(r *dto.EmployeeCreateRequest) { r.Salary = -1 }, "salary must be non-negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc()
			req := validEmployeeRequest()
			tt.modify(req)

			_, err := svc.Create(context.Background(), req)
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestEmployeeService_Create_InactiveCountry(t *testing.T) {
	svc := newSvc()
	req := validEmployeeRequest()
	req.CountryID = 2 // inactive in mock

	_, err := svc.Create(context.Background(), req)
	if err == nil || err.Error() != "country does not exist or is inactive" {
		t.Errorf("error = %v, want 'country does not exist or is inactive'", err)
	}
}

func TestEmployeeService_Create_NonExistentCountry(t *testing.T) {
	svc := newSvc()
	req := validEmployeeRequest()
	req.CountryID = 999

	_, err := svc.Create(context.Background(), req)
	if err == nil || err.Error() != "country does not exist or is inactive" {
		t.Errorf("error = %v", err)
	}
}

func TestEmployeeService_Create_InactiveJobTitle(t *testing.T) {
	svc := newSvc()
	req := validEmployeeRequest()
	req.JobTitleID = 2 // inactive

	_, err := svc.Create(context.Background(), req)
	if err == nil || err.Error() != "job title does not exist or is inactive" {
		t.Errorf("error = %v", err)
	}
}

func TestEmployeeService_Create_DuplicateEmail(t *testing.T) {
	svc := newSvc()

	if _, err := svc.Create(context.Background(), validEmployeeRequest()); err != nil {
		t.Fatalf("first Create error = %v", err)
	}
	dup := validEmployeeRequest()
	dup.FirstName = "Different"
	_, err := svc.Create(context.Background(), dup)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Errorf("error = %v, want ErrDuplicateEmail", err)
	}
}

func TestEmployeeService_Create_NormalisesInput(t *testing.T) {
	svc := newSvc()
	req := validEmployeeRequest()
	req.Email = "  JANE@Example.COM  "
	req.FirstName = "  Jane  "
	req.LastName = "  Doe  "
	req.Address = "  123 Main St  "

	resp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp.Email != "jane@example.com" {
		t.Errorf("Email = %q, want lowercase trimmed", resp.Email)
	}
	if resp.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want 'Jane Doe' (first/last must be trimmed before joining)", resp.FullName)
	}
	if resp.Address != "123 Main St" {
		t.Errorf("Address = %q", resp.Address)
	}
}

// =============================================================================
// GetByID
// =============================================================================

func TestEmployeeService_GetByID_Success(t *testing.T) {
	svc := newSvc()
	created, _ := svc.Create(context.Background(), validEmployeeRequest())

	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID error = %v", err)
	}
	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want Jane Doe", got.FullName)
	}
	if got.Email != "jane@example.com" {
		t.Errorf("Email = %q", got.Email)
	}
}

func TestEmployeeService_GetByID_InvalidID(t *testing.T) {
	svc := newSvc()
	for _, id := range []int64{0, -1} {
		_, err := svc.GetByID(context.Background(), id)
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("GetByID(%d) error = %v, want ErrInvalidInput", id, err)
		}
	}
}

func TestEmployeeService_GetByID_NotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.GetByID(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

// =============================================================================
// Update
// =============================================================================

func TestEmployeeService_Update_Success(t *testing.T) {
	svc := newSvc()
	created, _ := svc.Create(context.Background(), validEmployeeRequest())

	upd := validEmployeeRequest()
	upd.Salary = 130000
	resp, err := svc.Update(context.Background(), created.ID, upd)
	if err != nil {
		t.Fatalf("Update error = %v", err)
	}
	if resp.Salary != 130000 {
		t.Errorf("Salary = %f, want 130000", resp.Salary)
	}
}

func TestEmployeeService_Update_InvalidID(t *testing.T) {
	svc := newSvc()
	_, err := svc.Update(context.Background(), 0, validEmployeeRequest())
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestEmployeeService_Update_NotFound(t *testing.T) {
	svc := newSvc()
	_, err := svc.Update(context.Background(), 999, validEmployeeRequest())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

// =============================================================================
// Delete
// =============================================================================

func TestEmployeeService_Delete_Success(t *testing.T) {
	svc := newSvc()
	created, _ := svc.Create(context.Background(), validEmployeeRequest())

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete error = %v", err)
	}
	got, err := svc.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID after Delete error = %v", err)
	}
	if got == nil || got.IsActive {
		t.Error("after soft delete, employee should still be retrievable with IsActive=false")
	}
}

func TestEmployeeService_Delete_InvalidID(t *testing.T) {
	svc := newSvc()
	if err := svc.Delete(context.Background(), 0); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestEmployeeService_Delete_NotFound(t *testing.T) {
	svc := newSvc()
	if err := svc.Delete(context.Background(), 999); !errors.Is(err, ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}

// =============================================================================
// List
// =============================================================================

func TestEmployeeService_List_AllRecordsByDefault(t *testing.T) {
	svc := newSvc()
	for _, name := range []string{"A", "B", "C"} {
		req := validEmployeeRequest()
		req.Email = name + "@x.com"
		req.FirstName = name
		if _, err := svc.Create(context.Background(), req); err != nil {
			t.Fatalf("Create error: %v", err)
		}
	}

	result, err := svc.List(context.Background(), model.EmployeeFilter{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total = %d, want 3", result.Total)
	}
	if result.Limit != 0 || result.Offset != 0 {
		t.Errorf("default pagination = limit=%d offset=%d, want 0/0 (all records)", result.Limit, result.Offset)
	}
}

func TestEmployeeService_List_CapsExcessiveLimit(t *testing.T) {
	svc := newSvc()
	result, err := svc.List(context.Background(), model.EmployeeFilter{
		Pagination: model.Pagination{Limit: 5000, Offset: -10},
	})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if result.Limit > 100 {
		t.Errorf("Limit = %d, want capped at 100", result.Limit)
	}
	if result.Offset < 0 {
		t.Errorf("Offset = %d, want clamped to 0", result.Offset)
	}
}

func TestEmployeeService_List_ReturnsEmptySliceNotNil(t *testing.T) {
	svc := newSvc()
	result, err := svc.List(context.Background(), model.EmployeeFilter{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if result.Employees == nil {
		t.Error("Employees should be empty slice, not nil")
	}
}

// =============================================================================
// Insights — input validation
// =============================================================================

func TestEmployeeService_GetSalaryRange_EmptyCountry(t *testing.T) {
	svc := newSvc()
	if _, err := svc.GetSalaryRangeByCountry(context.Background(), ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestEmployeeService_GetSalaryByTitle_EmptyInputs(t *testing.T) {
	svc := newSvc()
	if _, err := svc.GetSalaryByTitle(context.Background(), "", "Engineer"); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v", err)
	}
	if _, err := svc.GetSalaryByTitle(context.Background(), "USA", ""); !errors.Is(err, ErrInvalidInput) {
		t.Errorf("error = %v", err)
	}
}

// =============================================================================
// model.Employee.FullName — tested here because FullName is the basis for
// EmployeeResponse.FullName in the converter.
// =============================================================================

func TestModelEmployee_FullName(t *testing.T) {
	tests := []struct {
		first, last, want string
	}{
		{"Jane", "Doe", "Jane Doe"},
		{"", "Doe", "Doe"},
		{"Jane", "", "Jane"},
	}
	for _, tt := range tests {
		e := model.Employee{FirstName: tt.first, LastName: tt.last}
		if got := e.FullName(); got != tt.want {
			t.Errorf("FullName(%q,%q) = %q, want %q", tt.first, tt.last, got, tt.want)
		}
	}
}
