package dto

import (
	"time"

	"github.com/salary-manager/backend/internal/model"
)

// CountryCreateRequest is the body of POST /api/countries.
type CountryCreateRequest struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Currency string `json:"currency"`
}

// CountryResponse is what the API returns for a single country. Today this
// shape is identical to model.Country, but it is a separate type so the
// public contract can evolve independently of the database schema.
type CountryResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Currency  string    `json:"currency"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CountryListResponse wraps a page of countries with pagination metadata.
type CountryListResponse struct {
	Countries []CountryResponse `json:"countries"`
	Total     int64             `json:"total"`
	Limit     int               `json:"limit"`
	Offset    int               `json:"offset"`
}

// ToCountryResponse converts a DB-layer Country into its API representation.
func ToCountryResponse(c *model.Country) *CountryResponse {
	if c == nil {
		return nil
	}
	return &CountryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Code:      c.Code,
		Currency:  c.Currency,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ToCountryListResponse converts a DB-layer list result into its API
// representation. The pagination fields pass through unchanged.
func ToCountryListResponse(r *model.CountryListResult) *CountryListResponse {
	if r == nil {
		return &CountryListResponse{Countries: []CountryResponse{}}
	}
	out := &CountryListResponse{
		Countries: make([]CountryResponse, 0, len(r.Countries)),
		Total:     r.Total,
		Limit:     r.Limit,
		Offset:    r.Offset,
	}
	for i := range r.Countries {
		out.Countries = append(out.Countries, *ToCountryResponse(&r.Countries[i]))
	}
	return out
}

// ToModelCountry builds a writable DB-layer Country value from a request.
// Normalisation (trim, uppercase) is the responsibility of the service or
// repository layers; this helper is a pure shape conversion.
func ToModelCountry(req *CountryCreateRequest) *model.Country {
	if req == nil {
		return &model.Country{}
	}
	return &model.Country{
		Name:     req.Name,
		Code:     req.Code,
		Currency: req.Currency,
	}
}
