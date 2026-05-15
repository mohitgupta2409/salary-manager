package service

import (
	"context"
	"strings"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type EmployeeService struct {
	repo         repository.EmployeeRepository
	countryRepo  repository.CountryRepository
	jobTitleRepo repository.JobTitleRepository
}

func NewEmployeeService(
	repo repository.EmployeeRepository,
	countryRepo repository.CountryRepository,
	jobTitleRepo repository.JobTitleRepository,
) *EmployeeService {
	return &EmployeeService{
		repo:         repo,
		countryRepo:  countryRepo,
		jobTitleRepo: jobTitleRepo,
	}
}

// Create accepts an API-layer request, validates it, persists the
// resulting model.Employee through the repository, and returns the
// API-layer response (with denormalised country/job-title fields).
func (s *EmployeeService) Create(ctx context.Context, req *dto.EmployeeCreateRequest) (*dto.EmployeeResponse, error) {
	if req == nil {
		return nil, ErrInvalidInput
	}
	emp := dto.ToModelEmployee(req)
	if err := validateEmployee(emp); err != nil {
		return nil, err
	}
	normalize(emp)

	if err := s.validateForeignKeys(ctx, emp); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, emp); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	return dto.ToEmployeeResponse(emp), nil
}

func (s *EmployeeService) GetByID(ctx context.Context, id int64) (*dto.EmployeeResponse, error) {
	if id <= 0 {
		return nil, ErrInvalidInput
	}
	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, ErrNotFound
	}
	return dto.ToEmployeeResponse(emp), nil
}

// Update applies the request fields to the row identified by id and
// returns the refreshed API view of the employee.
func (s *EmployeeService) Update(ctx context.Context, id int64, req *dto.EmployeeUpdateRequest) (*dto.EmployeeResponse, error) {
	if id <= 0 || req == nil {
		return nil, ErrInvalidInput
	}
	emp := dto.ToModelEmployee(req)
	emp.ID = id
	if err := validateEmployee(emp); err != nil {
		return nil, err
	}
	normalize(emp)

	if err := s.validateForeignKeys(ctx, emp); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, emp); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrNotFound
		}
		if isUniqueViolation(err) {
			return nil, ErrDuplicateEmail
		}
		return nil, err
	}
	return dto.ToEmployeeResponse(emp), nil
}

// Delete performs a soft delete via the repository (sets is_active = 0).
// The row is retained; subsequent List calls will not return it.
func (s *EmployeeService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// List forwards a filter (with embedded pagination) to the repository,
// applying default and safety bounds: a non-positive Limit is treated as
// "all records" by the repo, but the service caps callers at 100 to avoid
// pathologically large pages over the HTTP API.
func (s *EmployeeService) List(ctx context.Context, filter model.EmployeeFilter) (*dto.EmployeeListResponse, error) {
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Limit < 0 {
		filter.Limit = 0
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	out, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return dto.ToEmployeeListResponse(out), nil
}

func (s *EmployeeService) GetSalaryRangeByCountry(ctx context.Context, country string) (*model.SalaryRange, error) {
	if strings.TrimSpace(country) == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.GetSalaryRangeByCountry(ctx, country)
}

func (s *EmployeeService) GetSalaryByTitle(ctx context.Context, country, jobTitle string) (*model.SalaryByTitle, error) {
	if strings.TrimSpace(country) == "" || strings.TrimSpace(jobTitle) == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.GetSalaryByTitle(ctx, country, jobTitle)
}

func (s *EmployeeService) GetDepartmentStats(ctx context.Context, country string) ([]model.DepartmentStats, error) {
	return s.repo.GetDepartmentStats(ctx, country)
}

func (s *EmployeeService) GetOrgSummary(ctx context.Context) (*model.OrgSummary, error) {
	return s.repo.GetOrgSummary(ctx)
}

// validateForeignKeys ensures the referenced country and job title exist
// and are active. This catches bad FK values with a clean error before the
// DB raises a constraint violation.
func (s *EmployeeService) validateForeignKeys(ctx context.Context, emp *model.Employee) error {
	if s.countryRepo != nil {
		c, err := s.countryRepo.GetByID(ctx, emp.CountryID)
		if err != nil {
			return err
		}
		if c == nil || !c.IsActive {
			return ErrCountryInactive
		}
	}
	if s.jobTitleRepo != nil {
		jt, err := s.jobTitleRepo.GetByID(ctx, emp.JobTitleID)
		if err != nil {
			return err
		}
		if jt == nil || !jt.IsActive {
			return ErrJobTitleInactive
		}
	}
	return nil
}

func validateEmployee(emp *model.Employee) error {
	if strings.TrimSpace(emp.FirstName) == "" {
		return ErrFirstNameRequired
	}
	if strings.TrimSpace(emp.LastName) == "" {
		return ErrLastNameRequired
	}
	if strings.TrimSpace(emp.Email) == "" {
		return ErrEmailRequired
	}
	if !strings.Contains(emp.Email, "@") {
		return ErrEmailInvalid
	}
	if emp.JobTitleID <= 0 {
		return ErrJobTitleRequired
	}
	if emp.CountryID <= 0 {
		return ErrCountryRequired
	}
	if emp.Salary < 0 {
		return ErrSalaryNegative
	}
	return nil
}

func normalize(emp *model.Employee) {
	emp.Email = strings.ToLower(strings.TrimSpace(emp.Email))
	emp.FirstName = strings.TrimSpace(emp.FirstName)
	emp.LastName = strings.TrimSpace(emp.LastName)
	emp.Address = strings.TrimSpace(emp.Address)
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
