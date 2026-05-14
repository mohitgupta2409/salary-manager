package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/service"
)

type CountryHandler struct {
	svc *service.CountryService
}

func NewCountryHandler(svc *service.CountryService) *CountryHandler {
	return &CountryHandler{svc: svc}
}

func (h *CountryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	return r
}

func (h *CountryHandler) List(w http.ResponseWriter, r *http.Request) {
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	out, err := h.svc.List(r.Context(), includeInactive)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *CountryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c model.Country
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body"))
		return
	}
	if err := h.svc.Create(r.Context(), &c); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

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
	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	out, err := h.svc.List(r.Context(), includeInactive)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var d model.Department
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body"))
		return
	}
	if err := h.svc.Create(r.Context(), &d); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

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
	includeInactive := q.Get("include_inactive") == "true"
	var deptID int64
	if v := q.Get("department_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			deptID = id
		}
	}
	out, err := h.svc.List(r.Context(), deptID, includeInactive)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *JobTitleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var jt model.JobTitle
	if err := json.NewDecoder(r.Body).Decode(&jt); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body"))
		return
	}
	if err := h.svc.Create(r.Context(), &jt); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, jt)
}
