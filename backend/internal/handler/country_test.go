package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
)

// =============================================================================
// GET /api/countries
// =============================================================================

func TestCountryHandler_List_Success(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/countries", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var result dto.CountryListResponse
	decodeJSON(t, rec, &result)
	if len(result.Countries) != 1 {
		t.Errorf("got %d countries, want 1", len(result.Countries))
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}
	// Confirm DTO shape: enriched friendly fields are present, raw DB
	// metadata is not (the response uses dto.CountryResponse).
	if result.Countries[0].Code != "US" || result.Countries[0].Currency != "USD" {
		t.Errorf("unexpected country payload: %+v", result.Countries[0])
	}
}

func TestCountryHandler_List_DefaultPageSize(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/countries", nil)
	var result dto.CountryListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != DefaultPageSize {
		t.Errorf("Limit = %d, want DefaultPageSize=%d", result.Limit, DefaultPageSize)
	}
}

func TestCountryHandler_List_RespectsExplicitLimit(t *testing.T) {
	r := setupRouter()
	rec := do(r, "GET", "/api/countries?limit=5&offset=0", nil)
	var result dto.CountryListResponse
	decodeJSON(t, rec, &result)
	if result.Limit != 5 {
		t.Errorf("Limit = %d, want 5", result.Limit)
	}
}

func TestCountryHandler_List_IncludeInactiveQueryAccepted(t *testing.T) {
	// The mock repo ignores the IncludeInactive flag, but the handler must
	// still accept the query parameter without failing.
	r := setupRouter()
	rec := do(r, "GET", "/api/countries?include_inactive=true", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// =============================================================================
// POST /api/countries
// =============================================================================

func TestCountryHandler_Create_Success(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]string{"name": "Brazil", "code": "BR", "currency": "BRL"})
	rec := do(r, "POST", "/api/countries", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.CountryResponse
	decodeJSON(t, rec, &resp)
	if resp.ID == 0 || resp.Name != "Brazil" || resp.Code != "BR" || resp.Currency != "BRL" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestCountryHandler_Create_TrimsInputs(t *testing.T) {
	r := setupRouter()
	body, _ := json.Marshal(map[string]string{
		"name": "  France  ", "code": "  FR  ", "currency": "  EUR  ",
	})
	rec := do(r, "POST", "/api/countries", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201. body=%s", rec.Code, rec.Body.String())
	}
	var resp dto.CountryResponse
	decodeJSON(t, rec, &resp)
	if resp.Name != "France" || resp.Code != "FR" || resp.Currency != "EUR" {
		t.Errorf("inputs not trimmed: %+v", resp)
	}
}

func TestCountryHandler_Create_InvalidJSON(t *testing.T) {
	r := setupRouter()
	rec := do(r, "POST", "/api/countries", []byte("garbage"))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestCountryHandler_Create_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{"empty name", map[string]string{"name": "", "code": "BR", "currency": "BRL"}},
		{"empty code", map[string]string{"name": "Brazil", "code": "", "currency": "BRL"}},
		{"empty currency", map[string]string{"name": "Brazil", "code": "BR", "currency": ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupRouter()
			body, _ := json.Marshal(tt.body)
			rec := do(r, "POST", "/api/countries", body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400. body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}
