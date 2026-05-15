package model

// Pagination expresses an offset/limit window over a result set. It is
// embedded into per-entity list-request types so every paginated endpoint
// shares the same wire format and semantics.
//
//   - Limit  — maximum number of records to return. A non-positive value
//     (Limit <= 0) means "no limit"; the repository returns all matching
//     records. This is the canonical "fetch all" sentinel.
//   - Offset — number of records to skip from the start of the result set.
//     Negative values are clamped to zero.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// IsAll reports whether the request asks for every matching record (no
// limit). Repositories use this to decide whether to apply a LIMIT clause.
func (p Pagination) IsAll() bool { return p.Limit <= 0 }

// Normalized returns a copy with negative values clamped to zero. Repository
// implementations should call this before building SQL.
func (p Pagination) Normalized() Pagination {
	if p.Limit < 0 {
		p.Limit = 0
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p
}
