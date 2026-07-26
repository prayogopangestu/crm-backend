// Package pagination provides reusable request/response types for paginated
// list endpoints so handlers and usecases can exchange page metadata without
// redefining the same struct per feature.
package pagination

// PageRequest represents an inbound page query.
type PageRequest struct {
	Page  int `json:"page"  schema:"page"`
	Limit int `json:"limit" schema:"limit"`
}

// Normalized clamps the request to a sane default range and returns a copy.
// Use it inside usecases to enforce min/max page sizes.
func (r PageRequest) Normalized(defaultLimit, maxLimit int) PageRequest {
	page := r.Page
	if page < 1 {
		page = 1
	}
	limit := r.Limit
	if limit < 1 {
		limit = defaultLimit
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	return PageRequest{Page: page, Limit: limit}
}

// Offset returns the SQL/query offset implied by the page request.
func (r PageRequest) Offset() int {
	if r.Page < 1 {
		return 0
	}
	return (r.Page - 1) * r.Limit
}

// PageResponse is the standard paginated payload returned by list endpoints.
type PageResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	TotalPages int         `json:"totalPages"`
}

// NewPageResponse builds a PageResponse computing the total pages from a limit.
func NewPageResponse(data interface{}, total int64, page, limit int) PageResponse {
	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}
	return PageResponse{Data: data, Total: total, Page: page, TotalPages: totalPages}
}
