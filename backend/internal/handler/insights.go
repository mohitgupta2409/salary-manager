package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/salary-manager/backend/internal/service"
)

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
