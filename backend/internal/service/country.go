package service

import (
	"context"
	"strings"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type CountryService struct {
	repo repository.CountryRepository
}

func NewCountryService(repo repository.CountryRepository) *CountryService {
	return &CountryService{repo: repo}
}

// List forwards the request to the repository and converts the DB-layer
// result into an API response. Limit <= 0 means "all records"; the service
// caps a positive limit at 200 to bound page size.
func (s *CountryService) List(ctx context.Context, req model.CountryListRequest) (*dto.CountryListResponse, error) {
	if req.Limit > 200 {
		req.Limit = 200
	}
	out, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.ToCountryListResponse(out), nil
}

// Create normalises and validates the request, persists the resulting
// model.Country, and returns the API view.
func (s *CountryService) Create(ctx context.Context, req *dto.CountryCreateRequest) (*dto.CountryResponse, error) {
	if req == nil {
		return nil, ErrInvalidInput
	}
	c := dto.ToModelCountry(req)
	c.Name = strings.TrimSpace(c.Name)
	c.Code = strings.TrimSpace(c.Code)
	c.Currency = strings.TrimSpace(c.Currency)
	if c.Name == "" {
		return nil, ErrCountryNameRequired
	}
	if c.Code == "" {
		return nil, ErrCountryCodeRequired
	}
	if c.Currency == "" {
		return nil, ErrCurrencyRequired
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return dto.ToCountryResponse(c), nil
}
