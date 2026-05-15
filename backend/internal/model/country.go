package model

import "time"

type Country struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Currency  string    `json:"currency"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CountryListRequest is the input to CountryRepository.List. It carries
// pagination and optional filters. Pagination is embedded so its fields
// (limit, offset) appear at the top level of the JSON wire format.
type CountryListRequest struct {
	Pagination
	IncludeInactive bool `json:"include_inactive"`
}

// CountryListResult is a page of countries together with the total number
// of records that matched the filter (regardless of pagination).
type CountryListResult struct {
	Countries []Country `json:"countries"`
	Total     int64     `json:"total"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}
