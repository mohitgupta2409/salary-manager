package service

import (
	"context"
	"testing"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
)

func TestCountryService_Create_Success_ReturnsResponseDTO(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())

	resp, err := svc.Create(context.Background(), &dto.CountryCreateRequest{
		Name: "Germany", Code: "DE", Currency: "EUR",
	})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp == nil {
		t.Fatal("Create returned nil response")
	}
	if resp.ID == 0 {
		t.Error("ID should be assigned")
	}
	if resp.Name != "Germany" || resp.Code != "DE" || resp.Currency != "EUR" {
		t.Errorf("response fields wrong: %+v", resp)
	}
}

func TestCountryService_Create_TrimsInputs(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())
	resp, err := svc.Create(context.Background(), &dto.CountryCreateRequest{
		Name: "  France  ", Code: "  FR  ", Currency: "  EUR  ",
	})
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if resp.Name != "France" || resp.Code != "FR" || resp.Currency != "EUR" {
		t.Errorf("inputs not trimmed: %+v", resp)
	}
}

func TestCountryService_Create_NilRequest(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())
	if _, err := svc.Create(context.Background(), nil); err != ErrInvalidInput {
		t.Errorf("error = %v, want ErrInvalidInput", err)
	}
}

func TestCountryService_Create_Validation(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())

	tests := []struct {
		name    string
		req     dto.CountryCreateRequest
		wantErr string
	}{
		{"empty name", dto.CountryCreateRequest{Name: "", Code: "XX", Currency: "USD"}, "country name is required"},
		{"empty code", dto.CountryCreateRequest{Name: "X", Code: "", Currency: "USD"}, "country code is required"},
		{"empty currency", dto.CountryCreateRequest{Name: "X", Code: "XX", Currency: ""}, "currency is required"},
		{"all whitespace", dto.CountryCreateRequest{Name: "   ", Code: "XX", Currency: "USD"}, "country name is required"},
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

func TestCountryService_List_NeverNil(t *testing.T) {
	svc := NewCountryService(&mockCountryRepo{byID: map[int64]*model.Country{}})
	out, err := svc.List(context.Background(), model.CountryListRequest{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out == nil || out.Countries == nil {
		t.Error("List should return empty slice, not nil")
	}
}

func TestCountryService_List_ReturnsResponseDTOs(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())
	out, err := svc.List(context.Background(), model.CountryListRequest{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out.Total < 1 {
		t.Errorf("Total = %d, want >=1", out.Total)
	}
	// Verify DTO conversion: at least one of the seeded countries is present
	// with the friendly fields populated.
	var us *dto.CountryResponse
	for i := range out.Countries {
		if out.Countries[i].Code == "US" {
			c := out.Countries[i]
			us = &c
			break
		}
	}
	if us == nil {
		t.Fatal("expected to find seeded US country in response")
	}
	if us.Name != "United States" || us.Currency != "USD" {
		t.Errorf("DTO fields wrong: %+v", us)
	}
}

func TestCountryService_List_CapsLimit(t *testing.T) {
	svc := NewCountryService(newMockCountryRepo())
	// Service caps Limit at 200; the request still succeeds and returns data.
	out, err := svc.List(context.Background(), model.CountryListRequest{
		Pagination: model.Pagination{Limit: 5000},
	})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if out.Total == 0 {
		t.Error("expected non-zero total")
	}
}
