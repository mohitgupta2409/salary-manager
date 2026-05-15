package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
)

// =============================================================================
// GET /api/departments
// =============================================================================

func TestDepartmentHandler_List_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/departments", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var result dto.DepartmentListResponse
	decodeJSON(t, rec, &result)
	if len(result.Departments) != 1 {
		t.Errorf("got %d departments, want 1", len(result.Departments))
	}
	if result.Departments[0].Name != "Engineering" {
		t.Errorf("Name = %q, want Engineering", result.Departments[0].Name)
	}
}

func TestDepartmentHandler_List_DefaultPageSize(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/departments", nil)
	var result dto.DepartmentListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != DefaultPageSize {
		t.Errorf("Limit = %d, want DefaultPageSize=%d", result.Limit, DefaultPageSize)
	}
}

func TestDepartmentHandler_List_RespectsExplicitPagination(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/departments?limit=50&offset=10", nil)
	var result dto.DepartmentListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != 50 || result.Offset != 10 {
		t.Errorf("pagination = %d/%d, want 50/10", result.Limit, result.Offset)
	}
}

// =============================================================================
// POST /api/departments
// =============================================================================

func TestDepartmentHandler_Create_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": "Sales"})
	rec := do(r, "POST", "/api/departments", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.DepartmentResponse
	decodeJSON(t, rec, &resp)
	if resp.ID == 0 || resp.Name != "Sales" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if !resp.IsActive {
		t.Error("new department should be active")
	}
}

func TestDepartmentHandler_Create_TrimsName(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": "  Marketing  "})
	rec := do(r, "POST", "/api/departments", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.DepartmentResponse
	decodeJSON(t, rec, &resp)
	if resp.Name != "Marketing" {
		t.Errorf("Name = %q, want Marketing", resp.Name)
	}
}

func TestDepartmentHandler_Create_InvalidJSON(t *testing.T) {
	r := setupRouter()
	rec := do(r, "POST", "/api/departments", []byte("not json"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestDepartmentHandler_Create_EmptyName(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": ""})
	rec := do(r, "POST", "/api/departments", body)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400. body=%s", rec.Code, rec.Body.String())
	}
}

func TestDepartmentHandler_Create_WhitespaceOnlyName(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": "   "})
	rec := do(r, "POST", "/api/departments", body)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400. body=%s", rec.Code, rec.Body.String())
	}
}
