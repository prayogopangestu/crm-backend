package request

// SearchQuery represents the query string parameters for the global search endpoint.
type SearchQuery struct {
	Q string `json:"q"`
}
