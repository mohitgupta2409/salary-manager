package handler

import (
	"net/http"
	"testing"

	"github.com/salary-manager/backend/internal/model"
)

func TestInsightsHandler_SalaryRange_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/insights/salary-range?country=USA", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var sr model.SalaryRange
	decodeJSON(t, rec, &sr)
	if sr.Country != "USA" || sr.Average == 0 {
		t.Errorf("unexpected response: %+v", sr)
	}
}

func TestInsightsHandler_SalaryRange_MissingCountry(t *testing.T) {
	// Service rejects empty country with ErrInvalidInput -> 400.
	r := setupRouter()
	rec := do(r, "GET", "/api/insights/salary-range", nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestInsightsHandler_SalaryByTitle_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/insights/salary-by-title?country=USA&job_title=Engineer", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var sbt model.SalaryByTitle
	decodeJSON(t, rec, &sbt)
	if sbt.Country != "USA" || sbt.JobTitle != "Engineer" {
		t.Errorf("unexpected response: %+v", sbt)
	}
}

func TestInsightsHandler_SalaryByTitle_MissingParams(t *testing.T) {
	tests := []struct {
		name, query string
	}{
		{"missing country", "/api/insights/salary-by-title?job_title=Engineer"},
		{"missing title", "/api/insights/salary-by-title?country=USA"},
		{"both missing", "/api/insights/salary-by-title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter()
			rec := do(r, "GET", tt.query, nil)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestInsightsHandler_DepartmentStats_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/insights/department-stats", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d", rec.Code)
	}
	var stats []model.DepartmentStats
	decodeJSON(t, rec, &stats)
	if len(stats) == 0 {
		t.Error("expected at least one department stat")
	}
}

func TestInsightsHandler_Summary_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/insights/summary", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var summary model.OrgSummary
	decodeJSON(t, rec, &summary)
	if summary.TotalEmployees != 50 {
		t.Errorf("TotalEmployees = %d, want 50", summary.TotalEmployees)
	}
}
