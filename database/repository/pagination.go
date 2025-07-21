package repository

// PaginatedResult holds the data for a single page along with all pagination metadata.
// It's generic and can be used for any data type.
//
// NextPage and PreviousPage are pointers (*int) so they can be nil (and omitted from JSON output)
// when there isn't a next or previous page.
type PaginatedResult[T any] struct {
	Data         []T   `json:"data"`
	Page         int   `json:"page"`
	TotalRecords int64 `json:"total_records"`
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}

// MapPaginatedResult transforms a paginated result containing items of a source type (S)
// into a new result containing items of a destination type (D).
//
// It takes a source PaginatedResult and a mapper function that defines the conversion
// logic from an item of type S to an item of type D.
//
// Type Parameters:
//   - S: The source type (e.g., a database model like database.Post).
//   - D: The destination type (e.g., an API response DTO like PostResponse).
//
// The function returns a new PaginatedResult with the transformed data, while preserving
// all original pagination metadata (TotalRecords, CurrentPage, etc.).
func MapPaginatedResult[S any, D any](source *PaginatedResult[S], mapper func(S) D) *PaginatedResult[D] {
	mappedData := make([]D, len(source.Data))

	// Iterate over the source data and apply the mapper function
	for i, item := range source.Data {
		mappedData[i] = mapper(item)
	}

	return &PaginatedResult[D]{
		Data:         mappedData,
		TotalRecords: source.TotalRecords,
		CurrentPage:  source.CurrentPage,
		PageSize:     source.PageSize,
		TotalPages:   source.TotalPages,
		NextPage:     source.NextPage,
		PreviousPage: source.PreviousPage,
	}
}
