package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/salary-manager/backend/internal/model"
	"github.com/salary-manager/backend/internal/service"
)

type EmployeeHandler struct {
	svc *service.EmployeeService
}

func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func InsightsRoutes(svc *service.EmployeeService) chi.Router {
	r := chi.NewRouter()

	r.Get("/salary-range", func(w http.ResponseWriter, req *http.Request) {
		country := req.URL.Query().Get("country")
		result, err := svc.GetSalaryRangeByCountry(req.Context(), country)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	r.Get("/salary-by-title", func(w http.ResponseWriter, req *http.Request) {
		country := req.URL.Query().Get("country")
		jobTitle := req.URL.Query().Get("job_title")
		result, err := svc.GetSalaryByTitle(req.Context(), country, jobTitle)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	r.Get("/department-stats", func(w http.ResponseWriter, req *http.Request) {
		country := req.URL.Query().Get("country")
		result, err := svc.GetDepartmentStats(req.Context(), country)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	r.Get("/summary", func(w http.ResponseWriter, req *http.Request) {
		result, err := svc.GetOrgSummary(req.Context())
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	return r
}

func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var emp model.Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body"))
		return
	}

	if err := h.svc.Create(r.Context(), &emp); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, emp)
}

func (h *EmployeeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	emp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	var emp model.Employee
	if err := json.NewDecoder(r.Body).Decode(&emp); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid JSON body"))
		return
	}
	emp.ID = id

	if err := h.svc.Update(r.Context(), &emp); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, emp)
}

func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "employee deleted"})
}

func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := model.EmployeeFilter{
		Search:   r.URL.Query().Get("search"),
		Country:  r.URL.Query().Get("country"),
		JobTitle: r.URL.Query().Get("job_title"),
	}

	if p := r.URL.Query().Get("page"); p != "" {
		if page, err := strconv.Atoi(p); err == nil {
			filter.Page = page
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filter.Limit = limit
		}
	}

	result, err := h.svc.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func parseID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse(err.Error()))
	case errors.Is(err, service.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
	case errors.Is(err, service.ErrDuplicateEmail):
		writeJSON(w, http.StatusConflict, errorResponse(err.Error()))
	default:
		if err.Error() == "full name is required" || err.Error() == "email is required" ||
			err.Error() == "email must be valid" || err.Error() == "job title is required" ||
			err.Error() == "department is required" || err.Error() == "country is required" ||
			err.Error() == "salary must be non-negative" {
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
		log.Printf("internal error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("internal server error"))
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func errorResponse(msg string) map[string]string {
	return map[string]string{"error": msg}
}
