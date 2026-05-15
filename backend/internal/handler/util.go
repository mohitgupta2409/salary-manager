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

// DefaultPageSize is the page size used by collection endpoints when the
// caller does not supply an explicit ?limit= query parameter. Keeping the
// value in one place lets us tune the default without hunting through
// individual handlers.
const DefaultPageSize = 20

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse(err.Error()))
	case errors.Is(err, service.ErrInvalidInput):
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
	case errors.Is(err, service.ErrDuplicateEmail):
		writeJSON(w, http.StatusConflict, errorResponse(err.Error()))
	case service.IsValidationError(err):
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
	default:
		log.Printf("internal error: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("internal server error"))
	}
}

func parseID(r *http.Request) (int64, error) {
	idStr := chi.URLParam(r, "id")
	return strconv.ParseInt(idStr, 10, 64)
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

// parsePagination reads ?limit=, ?offset= and (optionally) ?page= from the
// request. When the caller omits ?limit= entirely we default to
// DefaultPageSize so collection endpoints never accidentally stream the
// whole table over the wire. A caller that genuinely wants every row must
// pass ?limit= with a non-positive value (e.g. limit=-1), which the
// repository treats as "no LIMIT".
func parsePagination(r *http.Request) model.Pagination {
	q := r.URL.Query()
	p := model.Pagination{Limit: DefaultPageSize}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			p.Offset = n
		}
	}
	if v := q.Get("page"); v != "" && p.Offset == 0 && p.Limit > 0 {
		if pg, err := strconv.Atoi(v); err == nil && pg > 1 {
			p.Offset = (pg - 1) * p.Limit
		}
	}
	return p
}
