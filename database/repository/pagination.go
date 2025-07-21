package repository

// PaginatedResult holds the data for a single page along with all pagination metadata.
// It's generic and can be used for any data type.
//
// NextPage and PreviousPage are pointers (*int) so they can be nil (and omitted from JSON output)
// when there isn't a next or previous page.
type PaginatedResult[T any] struct {
	Data         []T `json:"data"`
	Page         int
	TotalRecords int64 `json:"total_records"`
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}
