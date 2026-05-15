package service

import (
	"context"
	"strings"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/repository"
)

type JobTitleService struct {
	repo     repository.JobTitleRepository
	deptRepo repository.DepartmentRepository
}

func NewJobTitleService(repo repository.JobTitleRepository, deptRepo repository.DepartmentRepository) *JobTitleService {
	return &JobTitleService{repo: repo, deptRepo: deptRepo}
}

func (s *JobTitleService) List(ctx context.Context, req model.JobTitleListRequest) (*dto.JobTitleListResponse, error) {
	if req.Limit > 200 {
		req.Limit = 200
	}
	out, err := s.repo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return dto.ToJobTitleListResponse(out), nil
}

// Create validates the department FK before persisting so that callers see
// a clean error before SQLite raises a constraint violation.
func (s *JobTitleService) Create(ctx context.Context, req *dto.JobTitleCreateRequest) (*dto.JobTitleResponse, error) {
	if req == nil {
		return nil, ErrInvalidInput
	}
	jt := dto.ToModelJobTitle(req)
	jt.Name = strings.TrimSpace(jt.Name)
	if jt.Name == "" {
		return nil, ErrJobTitleNameRequired
	}
	if jt.DepartmentID <= 0 {
		return nil, ErrDepartmentRequired
	}
	if s.deptRepo != nil {
		d, err := s.deptRepo.GetByID(ctx, jt.DepartmentID)
		if err != nil {
			return nil, err
		}
		if d == nil || !d.IsActive {
			return nil, ErrDepartmentInactive
		}
	}
	if err := s.repo.Create(ctx, jt); err != nil {
		return nil, err
	}
	// repo.Create does not denormalize; populate the Department name on the
	// response from the lookup we already performed above.
	resp := dto.ToJobTitleResponse(jt)
	if s.deptRepo != nil {
		if d, err := s.deptRepo.GetByID(ctx, jt.DepartmentID); err == nil && d != nil {
			resp.Department = d.Name
		}
	}
	return resp, nil
}
