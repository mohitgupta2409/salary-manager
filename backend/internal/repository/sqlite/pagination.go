package sqlite

import "github.com/salary-manager/backend/internal/model"

// applyPagination appends LIMIT/OFFSET clauses to query and the corresponding
// arguments to args. When p.IsAll() is true (Limit <= 0), the query is left
// unchanged so all matching rows are returned. Negative offsets are clamped
// to zero. Returns the new query and args.
//
// Use this helper from every repository's List method to keep pagination
// semantics consistent across entities.
func applyPagination(query string, args []interface{}, p model.Pagination) (string, []interface{}) {
	p = p.Normalized()
	if p.IsAll() {
		if p.Offset > 0 {
			query += " LIMIT -1 OFFSET ?"
			args = append(args, p.Offset)
		}
		return query, args
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, p.Limit, p.Offset)
	return query, args
}
