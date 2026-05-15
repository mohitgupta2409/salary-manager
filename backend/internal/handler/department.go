package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/salary-manager/backend/internal/dto"
	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/service"
)

type DepartmentHandler struct {
	svc *service.DepartmentService
}

func NewDepartmentHandler(svc *service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{svc: svc}
}

func (h *DepartmentHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	return r
}

func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	req := model.DepartmentListRequest{
		Pagination:      parsePagination(r),
		IncludeInactive: r.URL.Query().Get("include_inactive") == "true",
	}
	resp, err := h.svc.List(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.DepartmentCreateRequest
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
