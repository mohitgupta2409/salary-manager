package service

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

// mockDeptRepo for service tests
type mockDeptRepo struct {
	byID map[int64]*model.Department
}

func newMockDeptRepo() *mockDeptRepo {
	return &mockDeptRepo{
		byID: map[int64]*model.Department{
			1: {ID: 1, Name: "Engineering", IsActive: true},
			2: {ID: 2, Name: "Inactive Dept", IsActive: false},
		},
	}
}

func (m *mockDeptRepo) Create(_ context.Context, d *model.Department) error {
	d.ID = int64(len(m.byID) + 1)
	m.byID[d.ID] = d
	return nil
}
func (m *mockDeptRepo) List(_ context.Context, _ bool) ([]model.Department, error) {
	out := []model.Department{}
	for _, d := range m.byID {
		out = append(out, *d)
	}
	return out, nil
}
func (m *mockDeptRepo) GetByID(_ context.Context, id int64) (*model.Department, error) {
	d, ok := m.byID[id]
	if !ok {
		return nil, nil
	}
	return d, nil
}

func TestCountryService_Create_Validation(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())

	tests := []struct {
		name    string
		country model.Country
		wantErr string
	}{
		{"empty name", model.Country{Name: "", Code: "XX", Currency: "USD"}, "country name is required"},
		{"empty code", model.Country{Name: "X", Code: "", Currency: "USD"}, "country code is required"},
		{"empty currency", model.Country{Name: "X", Code: "XX", Currency: ""}, "currency is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(context.Background(), &tt.country)
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestCountryService_List_NeverNil(t *testing.T) {
	svc := NewCountryService(&mockCountryRepo{byID: map[int64]*model.Country{}})
	out, err := svc.List(context.Background(), false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if out == nil {
		t.Error("List() should return empty slice, not nil")
	}
}

func TestDepartmentService_Create_Validation(t *testing.T) {
	svc := NewDepartmentService(newMockDeptRepo())
	err := svc.Create(context.Background(), &model.Department{Name: ""})
	if err == nil || err.Error() != "department name is required" {
		t.Errorf("error = %v", err)
	}
}

func TestDepartmentService_List_NeverNil(t *testing.T) {
	svc := NewDepartmentService(&mockDeptRepo{byID: map[int64]*model.Department{}})
	out, err := svc.List(context.Background(), false)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if out == nil {
		t.Error("List() should return empty slice, not nil")
	}
}

func TestJobTitleService_Create_Validation(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())

	tests := []struct {
		name    string
		jt      model.JobTitle
		wantErr string
	}{
		{"empty name", model.JobTitle{Name: "", DepartmentID: 1}, "job title name is required"},
		{"missing dept", model.JobTitle{Name: "Engineer", DepartmentID: 0}, "department is required"},
		{"non-existent dept", model.JobTitle{Name: "Engineer", DepartmentID: 999}, "department does not exist or is inactive"},
		{"inactive dept", model.JobTitle{Name: "Engineer", DepartmentID: 2}, "department does not exist or is inactive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Create(context.Background(), &tt.jt)
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestJobTitleService_Create_Success(t *testing.T) {
	svc := NewJobTitleService(newMockJobTitleRepo(), newMockDeptRepo())
	jt := &model.JobTitle{Name: "DevOps Engineer", DepartmentID: 1}
	if err := svc.Create(context.Background(), jt); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if jt.ID == 0 {
		t.Error("Create() should set ID")
	}
}
