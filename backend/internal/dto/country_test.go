package dto

import (
	"testing"
	"time"

	"github.com/salary-manager/backend/internal/model"
)

func TestToCountryResponse(t *testing.T) {
	now := time.Now().UTC()
	got := ToCountryResponse(&model.Country{
		ID: 7, Name: "Germany", Code: "DE", Currency: "EUR",
		IsActive: true, CreatedAt: now, UpdatedAt: now,
	})
	if got == nil {
		t.Fatal("got nil")
	}
	if got.ID != 7 || got.Name != "Germany" || got.Code != "DE" || got.Currency != "EUR" {
		t.Errorf("response wrong: %+v", got)
	}
	if !got.IsActive || !got.CreatedAt.Equal(now) {
		t.Errorf("metadata wrong: %+v", got)
	}
}

func TestToCountryResponse_Nil(t *testing.T) {
	if got := ToCountryResponse(nil); got != nil {
		t.Errorf("nil input should yield nil, got %+v", got)
	}
}

func TestToCountryListResponse(t *testing.T) {
	in := &model.CountryListResult{
		Countries: []model.Country{
			{ID: 1, Name: "USA", Code: "US", Currency: "USD"},
			{ID: 2, Name: "Germany", Code: "DE", Currency: "EUR"},
		},
		Total: 2, Limit: 10, Offset: 0,
	}
	got := ToCountryListResponse(in)
	if len(got.Countries) != 2 {
		t.Fatalf("len = %d, want 2", len(got.Countries))
	}
	if got.Total != 2 || got.Limit != 10 || got.Offset != 0 {
		t.Errorf("pagination metadata mismatch: %+v", got)
	}
	if got.Countries[0].Code != "US" || got.Countries[1].Code != "DE" {
		t.Errorf("order or content mismatch: %+v", got.Countries)
	}
}

func TestToCountryListResponse_Nil(t *testing.T) {
	got := ToCountryListResponse(nil)
	if got == nil || got.Countries == nil {
		t.Error("expected non-nil response with empty slice")
	}
}

func TestToModelCountry(t *testing.T) {
	got := ToModelCountry(&CountryCreateRequest{
		Name: "Brazil", Code: "BR", Currency: "BRL",
	})
	if got.Name != "Brazil" || got.Code != "BR" || got.Currency != "BRL" {
		t.Errorf("conversion wrong: %+v", got)
	}
	if got.ID != 0 {
		t.Errorf("ID should be zero on a request, got %d", got.ID)
	}
}

func TestToModelCountry_Nil(t *testing.T) {
	got := ToModelCountry(nil)
	if got == nil {
		t.Error("expected zero-value model, got nil")
	}
}
