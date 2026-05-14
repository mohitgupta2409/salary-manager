package service

import (
	"context"
	"errors"
	"strings"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotFound       = errors.New("employee not found")
	ErrDuplicateEmail = errors.New("email already exists")
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

func (s *EmployeeService) Create(ctx context.Context, emp *model.Employee) error {
	if err := validateEmployee(emp); err != nil {
		return err
	}
	normalize(emp)

	if err := s.validateForeignKeys(ctx, emp); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, emp); err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (s *EmployeeService) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
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
	return emp, nil
}

func (s *EmployeeService) Update(ctx context.Context, emp *model.Employee) error {
	if emp.ID <= 0 {
		return ErrInvalidInput
	}
	if err := validateEmployee(emp); err != nil {
		return err
	}
	normalize(emp)

	if err := s.validateForeignKeys(ctx, emp); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, emp); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrNotFound
		}
		if isUniqueViolation(err) {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

// Delete performs a soft delete (sets is_active = 0).
func (s *EmployeeService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *EmployeeService) List(ctx context.Context, filter model.EmployeeFilter) (*model.EmployeeListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}
	return s.repo.List(ctx, filter)
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
			return errors.New("country does not exist or is inactive")
		}
	}
	if s.jobTitleRepo != nil {
		jt, err := s.jobTitleRepo.GetByID(ctx, emp.JobTitleID)
		if err != nil {
			return err
		}
		if jt == nil || !jt.IsActive {
			return errors.New("job title does not exist or is inactive")
		}
	}
	return nil
}

func validateEmployee(emp *model.Employee) error {
	if strings.TrimSpace(emp.FirstName) == "" {
		return errors.New("first name is required")
	}
	if strings.TrimSpace(emp.LastName) == "" {
		return errors.New("last name is required")
	}
	if strings.TrimSpace(emp.Email) == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(emp.Email, "@") {
		return errors.New("email must be valid")
	}
	if emp.JobTitleID <= 0 {
		return errors.New("job title is required")
	}
	if emp.CountryID <= 0 {
		return errors.New("country is required")
	}
	if emp.Salary < 0 {
		return errors.New("salary must be non-negative")
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
