package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/service"
)

type JobTitleHandler struct {
	svc *service.JobTitleService
}

func NewJobTitleHandler(svc *service.JobTitleService) *JobTitleHandler {
	return &JobTitleHandler{svc: svc}
}

func (h *JobTitleHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	return r
}

func (h *JobTitleHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := model.JobTitleListRequest{
		Pagination:      parsePagination(r),
		IncludeInactive: q.Get("include_inactive") == "true",
	}
	if v := q.Get("department_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.DepartmentID = id
		}
	}
	resp, err := h.svc.List(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *JobTitleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.JobTitleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body"))
		return
	}
	resp, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}
