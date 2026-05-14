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
	q := r.URL.Query()
	filter := model.EmployeeFilter{Search: q.Get("search")}

	if v := q.Get("country_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.CountryID = id
		}
	}
	if v := q.Get("job_title_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.JobTitleID = id
		}
	}
	if v := q.Get("department_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.DepartmentID = id
		}
	}
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Page = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
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
		if isValidationError(err.Error()) {
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
		log.Printf("internal error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("internal server error"))
	}
}

func isValidationError(msg string) bool {
	switch msg {
	case "first name is required", "last name is required",
		"email is required", "email must be valid",
		"job title is required", "country is required",
		"department is required", "salary must be non-negative",
		"country does not exist or is inactive",
		"job title does not exist or is inactive",
		"department does not exist or is inactive",
		"country name is required", "country code is required",
		"currency is required", "department name is required",
		"job title name is required":
		return true
	}
	return false
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
