package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/dto"
)

// =============================================================================
// POST /api/employees
// =============================================================================

func TestEmployeeHandler_Create_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "POST", "/api/employees", validEmployeeJSON())

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.EmployeeResponse
	decodeJSON(t, rec, &resp)
	if resp.ID == 0 || resp.FullName != "Jane Doe" {
		t.Errorf("unexpected response: %+v", resp)
	}
	// Confirm the API contract: enriched country/job_title surface, raw
	// FK ids do not.
	if resp.Country == "" || resp.JobTitle == "" {
		t.Errorf("denormalised fields missing: %+v", resp)
	}
}

func TestEmployeeHandler_Create_InvalidJSON(t *testing.T) {
	r := setupRouter()
	rec := do(r, "POST", "/api/employees", []byte("not json"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestEmployeeHandler_Create_ValidationError(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]interface{}{"first_name": ""})
	rec := do(r, "POST", "/api/employees", body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400. body=%s", rec.Code, rec.Body.String())
	}
}

func TestEmployeeHandler_Create_DuplicateEmail(t *testing.T) {
	r := setupRouter()
	if rec := do(r, "POST", "/api/employees", validEmployeeJSON()); rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d", rec.Code)
	}
	rec := do(r, "POST", "/api/employees", validEmployeeJSON())
	if rec.Code != http.StatusConflict {
		t.Errorf("second create status = %d, want 409", rec.Code)
	}
}

// =============================================================================
// GET /api/employees/{id}
// =============================================================================

func TestEmployeeHandler_GetByID_Success(t *testing.T) {
	r := setupRouter()
	if rec := do(r, "POST", "/api/employees", validEmployeeJSON()); rec.Code != http.StatusCreated {
		t.Fatalf("seed create failed: %d", rec.Code)
	}
	rec := do(r, "GET", "/api/employees/1", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	var resp dto.EmployeeResponse
	decodeJSON(t, rec, &resp)
	if resp.ID != 1 {
		t.Errorf("ID = %d, want 1", resp.ID)
	}
}

func TestEmployeeHandler_GetByID_NotFound(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/employees/999", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestEmployeeHandler_GetByID_InvalidID(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/employees/abc", nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// =============================================================================
// PUT /api/employees/{id}
// =============================================================================

func TestEmployeeHandler_Update_Success(t *testing.T) {
	r := setupRouter()
	if rec := do(r, "POST", "/api/employees", validEmployeeJSON()); rec.Code != http.StatusCreated {
		t.Fatalf("seed create failed: %d", rec.Code)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"first_name": "Jane", "last_name": "Smith", "email": "jane@example.com",
		"job_title_id": 1, "country_id": 1, "salary": 130000,
		"join_date": time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	})
	rec := do(r, "PUT", "/api/employees/1", body)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.EmployeeResponse
	decodeJSON(t, rec, &resp)
	if resp.FullName != "Jane Smith" || resp.Salary != 130000 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestEmployeeHandler_Update_NotFound(t *testing.T) {
	r := setupRouter()
	rec := do(r, "PUT", "/api/employees/999", validEmployeeJSON())
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestEmployeeHandler_Update_InvalidJSON(t *testing.T) {
	r := setupRouter()
	rec := do(r, "PUT", "/api/employees/1", []byte("garbage"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// =============================================================================
// DELETE /api/employees/{id}
// =============================================================================

func TestEmployeeHandler_Delete_Success(t *testing.T) {
	r := setupRouter()
	if rec := do(r, "POST", "/api/employees", validEmployeeJSON()); rec.Code != http.StatusCreated {
		t.Fatalf("seed create failed: %d", rec.Code)
	}
	rec := do(r, "DELETE", "/api/employees/1", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("delete status = %d", rec.Code)
	}

	rec = do(r, "GET", "/api/employees", nil)
	var result dto.EmployeeListResponse
	decodeJSON(t, rec, &result)
	if result.Total != 0 {
		t.Errorf("after soft delete, list total = %d, want 0", result.Total)
	}
}

func TestEmployeeHandler_Delete_NotFound(t *testing.T) {
	r := setupRouter()
	rec := do(r, "DELETE", "/api/employees/999", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

// =============================================================================
// GET /api/employees
// =============================================================================

func TestEmployeeHandler_List_Success(t *testing.T) {
	r := setupRouter()
	if rec := do(r, "POST", "/api/employees", validEmployeeJSON()); rec.Code != http.StatusCreated {
		t.Fatalf("seed create failed: %d", rec.Code)
	}
	rec := do(r, "GET", "/api/employees?page=1&limit=10", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	var result dto.EmployeeListResponse
	decodeJSON(t, rec, &result)
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if result.Limit != 10 {
		t.Errorf("Limit = %d, want 10", result.Limit)
	}
}

func TestEmployeeHandler_List_DefaultPageSize(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/employees", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var result dto.EmployeeListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != DefaultPageSize {
		t.Errorf("Limit = %d, want DefaultPageSize=%d", result.Limit, DefaultPageSize)
	}
}
