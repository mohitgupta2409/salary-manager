package service

import (
	"context"
	"errors"
	"strings"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type CountryService struct {
	repo repository.CountryRepository
}

func NewCountryService(repo repository.CountryRepository) *CountryService {
	return &CountryService{repo: repo}
}

func (s *CountryService) List(ctx context.Context, includeInactive bool) ([]model.Country, error) {
	out, err := s.repo.List(ctx, includeInactive)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Country{}
	}
	return out, nil
}

func (s *CountryService) Create(ctx context.Context, c *model.Country) error {
	c.Name = strings.TrimSpace(c.Name)
	c.Code = strings.TrimSpace(c.Code)
	c.Currency = strings.TrimSpace(c.Currency)
	if c.Name == "" {
		return errors.New("country name is required")
	}
	if c.Code == "" {
		return errors.New("country code is required")
	}
	if c.Currency == "" {
		return errors.New("currency is required")
	}
	return s.repo.Create(ctx, c)
}

type DepartmentService struct {
	repo repository.DepartmentRepository
}

func NewDepartmentService(repo repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

func (s *DepartmentService) List(ctx context.Context, includeInactive bool) ([]model.Department, error) {
	out, err := s.repo.List(ctx, includeInactive)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.Department{}
	}
	return out, nil
}

func (s *DepartmentService) Create(ctx context.Context, d *model.Department) error {
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return errors.New("department name is required")
	}
	return s.repo.Create(ctx, d)
}

type JobTitleService struct {
	repo     repository.JobTitleRepository
	deptRepo repository.DepartmentRepository
}

func NewJobTitleService(repo repository.JobTitleRepository, deptRepo repository.DepartmentRepository) *JobTitleService {
	return &JobTitleService{repo: repo, deptRepo: deptRepo}
}

func (s *JobTitleService) List(ctx context.Context, departmentID int64, includeInactive bool) ([]model.JobTitle, error) {
	out, err := s.repo.List(ctx, departmentID, includeInactive)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []model.JobTitle{}
	}
	return out, nil
}

func (s *JobTitleService) Create(ctx context.Context, jt *model.JobTitle) error {
	jt.Name = strings.TrimSpace(jt.Name)
	if jt.Name == "" {
		return errors.New("job title name is required")
	}
	if jt.DepartmentID <= 0 {
		return errors.New("department is required")
	}
	if s.deptRepo != nil {
		d, err := s.deptRepo.GetByID(ctx, jt.DepartmentID)
		if err != nil {
			return err
		}
		if d == nil || !d.IsActive {
			return errors.New("department does not exist or is inactive")
		}
	}
	return s.repo.Create(ctx, jt)
}
