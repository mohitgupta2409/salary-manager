package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

func TestToEmployeeResponse_FullEnrichment(t *testing.T) {
	join := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	now := time.Now().UTC()
	in := &model.Employee{
		ID: 42, FirstName: "Jane", LastName: "Doe", Email: "jane@x.com",
		JobTitleID: 7, CountryID: 9, Salary: 100000,
		Address: "1 Way", JoinDate: join, IsActive: true,
		CreatedAt: now, UpdatedAt: now,
		Country: "Germany", Currency: "EUR",
		JobTitle: "Engineer", Department: "Platform",
	}
	got := ToEmployeeResponse(in)
	if got == nil {
		t.Fatal("got nil")
	}
	if got.ID != 42 || got.Email != "jane@x.com" {
		t.Errorf("identity fields wrong: %+v", got)
	}
	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q, want 'Jane Doe' (FirstName/LastName must collapse into FullName)", got.FullName)
	}
	if got.Country != "Germany" {
		t.Errorf("Country = %q, want Germany", got.Country)
	}
	if got.JobTitle != "Engineer" || got.Department != "Platform" {
		t.Errorf("job title/department denorm wrong: %+v", got)
	}
	if got.Currency != "EUR" {
		t.Errorf("Currency = %q, want EUR (needed so the UI can format Salary in the local currency)", got.Currency)
	}
	if !got.JoinDate.Equal(join) || got.Salary != 100000 {
		t.Errorf("misc fields wrong: %+v", got)
	}
}

func TestToEmployeeResponse_Nil(t *testing.T) {
	if got := ToEmployeeResponse(nil); got != nil {
		t.Errorf("nil input should yield nil, got %+v", got)
	}
}

func TestToEmployeeListResponse(t *testing.T) {
	in := &model.EmployeeListResult{
		Employees: []model.Employee{
			{ID: 1, FirstName: "A", LastName: "B", Country: "USA"},
			{ID: 2, FirstName: "C", LastName: "D", Country: "Germany"},
		},
		Total: 2, Limit: 10, Offset: 0,
	}
	got := ToEmployeeListResponse(in)
	if len(got.Employees) != 2 {
		t.Fatalf("len = %d, want 2", len(got.Employees))
	}
	if got.Employees[0].FullName != "A B" || got.Employees[0].Country != "USA" {
		t.Errorf("employee[0] = %+v", got.Employees[0])
	}
	if got.Total != 2 || got.Limit != 10 {
		t.Errorf("pagination metadata mismatch: %+v", got)
	}
}

// TestEmployeeResponse_HidesDBOnlyFields documents that FK ids and raw
// first/last names are NOT part of the API response, while denormalised
// display fields (including currency, used by the UI to format salary)
// are. This is enforced by the struct definition, so the test pins the
// JSON shape.
func TestEmployeeResponse_HidesDBOnlyFields(t *testing.T) {
	r := ToEmployeeResponse(&model.Employee{
		ID: 1, FirstName: "Jane", LastName: "Doe", Email: "j@x.com",
		CountryID: 9, JobTitleID: 7, Currency: "EUR",
		Country: "Germany", JobTitle: "Engineer", Department: "Platform",
	})
	for _, hidden := range []string{"first_name", "last_name", "country_id", "job_title_id"} {
		if jsonHasField(t, r, hidden) {
			t.Errorf("EmployeeResponse JSON unexpectedly contains %q", hidden)
		}
	}
	for _, exposed := range []string{"id", "full_name", "email", "country", "currency", "job_title", "department"} {
		if !jsonHasField(t, r, exposed) {
			t.Errorf("EmployeeResponse JSON missing required field %q", exposed)
		}
	}
}

func jsonHasField(t *testing.T, v interface{}, field string) bool {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	_, ok := m[field]
	return ok
}

func TestToEmployeeListResponse_Nil(t *testing.T) {
	got := ToEmployeeListResponse(nil)
	if got == nil || got.Employees == nil {
		t.Error("expected non-nil response with empty slice")
	}
}

func TestToModelEmployee(t *testing.T) {
	join := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	got := ToModelEmployee(&EmployeeCreateRequest{
		FirstName: "Jane", LastName: "Doe", Email: "j@x.com",
		JobTitleID: 1, CountryID: 2, Salary: 50000,
		Address: "addr", JoinDate: join,
	})
	if got.FirstName != "Jane" || got.JobTitleID != 1 || got.CountryID != 2 {
		t.Errorf("conversion wrong: %+v", got)
	}
	if !got.JoinDate.Equal(join) {
		t.Errorf("JoinDate not preserved")
	}
	if got.IsActive {
		t.Error("IsActive should default to false (repo sets it to true on insert)")
	}
}

func TestToModelEmployee_Nil(t *testing.T) {
	if got := ToModelEmployee(nil); got == nil {
		t.Error("expected zero-value model, got nil")
	}
}
