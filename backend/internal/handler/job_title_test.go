package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
)

// =============================================================================
// GET /api/job-titles
// =============================================================================

func TestJobTitleHandler_List_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/job-titles", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var result dto.JobTitleListResponse
	decodeJSON(t, rec, &result)
	if len(result.JobTitles) != 1 {
		t.Errorf("got %d job titles, want 1", len(result.JobTitles))
	}
	// Confirm the denormalised Department name is surfaced.
	if result.JobTitles[0].Department != "Engineering" {
		t.Errorf("Department = %q, want Engineering", result.JobTitles[0].Department)
	}
}

func TestJobTitleHandler_List_DefaultPageSize(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/job-titles", nil)
	var result dto.JobTitleListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != DefaultPageSize {
		t.Errorf("Limit = %d, want DefaultPageSize=%d", result.Limit, DefaultPageSize)
	}
}

func TestJobTitleHandler_List_FilterByDepartment(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/job-titles?department_id=1", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	// The mock repo doesn't actually filter, but the handler must accept
	// the parameter and forward it without erroring.
	var result dto.JobTitleListResponse
	decodeJSON(t, rec, &result)
	if result.Total < 1 {
		t.Errorf("Total = %d, want >=1", result.Total)
	}
}

func TestJobTitleHandler_List_IgnoresMalformedDepartmentID(t *testing.T) {
	// Non-numeric department_id should be quietly ignored, not 500.
	r := setupRouter()
	rec := do(r, "GET", "/api/job-titles?department_id=abc", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestJobTitleHandler_List_RespectsExplicitPagination(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/job-titles?limit=25&offset=5", nil)
	var result dto.JobTitleListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != 25 || result.Offset != 5 {
		t.Errorf("pagination = %d/%d, want 25/5", result.Limit, result.Offset)
	}
}

// =============================================================================
// POST /api/job-titles
// =============================================================================

func TestJobTitleHandler_Create_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]interface{}{
		"name": "DevOps Engineer", "department_id": 1,
	})
	rec := do(r, "POST", "/api/job-titles", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.JobTitleResponse
	decodeJSON(t, rec, &resp)
	if resp.ID == 0 || resp.Name != "DevOps Engineer" {
		t.Errorf("unexpected response: %+v", resp)
	}
	// Service is responsible for populating Department on the create
	// response (the repo Create only writes raw columns).
	if resp.Department != "Engineering" {
		t.Errorf("Department = %q, want Engineering", resp.Department)
	}
}

func TestJobTitleHandler_Create_InvalidJSON(t *testing.T) {
	r := setupRouter()
	rec := do(r, "POST", "/api/job-titles", []byte("not json"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestJobTitleHandler_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"empty name", map[string]interface{}{"name": "", "department_id": 1}},
		{"missing department", map[string]interface{}{"name": "Engineer"}},
		{"non-existent department", map[string]interface{}{"name": "Engineer", "department_id": 999}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter()
			body, _ := json.Marshal(tt.body)
			rec := do(r, "POST", "/api/job-titles", body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400. body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
