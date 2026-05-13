package service

import (
	"context"
	"errors"
	"strings"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrNotFound      = errors.New("employee not found")
	ErrDuplicateEmail = errors.New("email already exists")
)

type EmployeeService struct {
	repo repository.EmployeeRepository
}

func NewEmployeeService(repo repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) Create(ctx context.Context, emp *model.Employee) error {
	if err := validateEmployee(emp); err != nil {
		return err
	}
	emp.Email = strings.ToLower(strings.TrimSpace(emp.Email))
	emp.FullName = strings.TrimSpace(emp.FullName)

	if err := s.repo.Create(ctx, emp); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
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
	emp.Email = strings.ToLower(strings.TrimSpace(emp.Email))
	emp.FullName = strings.TrimSpace(emp.FullName)

	if err := s.repo.Update(ctx, emp); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrNotFound
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

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

func validateEmployee(emp *model.Employee) error {
	if strings.TrimSpace(emp.FullName) == "" {
		return errors.New("full name is required")
	}
	if strings.TrimSpace(emp.Email) == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(emp.Email, "@") {
		return errors.New("email must be valid")
	}
	if strings.TrimSpace(emp.JobTitle) == "" {
		return errors.New("job title is required")
	}
	if strings.TrimSpace(emp.Department) == "" {
		return errors.New("department is required")
	}
	if strings.TrimSpace(emp.Country) == "" {
		return errors.New("country is required")
	}
	if emp.Salary < 0 {
		return errors.New("salary must be non-negative")
	}
	if strings.TrimSpace(emp.Currency) == "" {
		emp.Currency = "USD"
	}
	return nil
}
