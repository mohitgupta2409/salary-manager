package service

import (
	"context"
	"strings"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type DepartmentService struct {
	repo repository.DepartmentRepository
}

func NewDepartmentService(repo repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{repo: repo}
}

// List forwards the request to the repository and converts the DB-layer
// result into an API response.
func (s *DepartmentService) List(ctx context.Context, req model.DepartmentListRequest) (*dto.DepartmentListResponse, error) {
	if req.Limit > 200 {
		req.Limit = 200
	}
	out, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.ToDepartmentListResponse(out), nil
}

func (s *DepartmentService) Create(ctx context.Context, req *dto.DepartmentCreateRequest) (*dto.DepartmentResponse, error) {
	if req == nil {
		return nil, ErrInvalidInput
	}
	d := dto.ToModelDepartment(req)
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		return nil, ErrDepartmentNameRequired
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return dto.ToDepartmentResponse(d), nil
}
