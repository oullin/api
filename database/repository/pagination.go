package repository

import "math"

const MaxLimit = 100

// Pagination holds the data for a single page along with all pagination metadata.
// It's generic and can be used for any data type.
//
// NextPage and PreviousPage are pointers (*int) so they can be nil (and omitted from JSON output)
// when there isn't a next or previous page.
type Pagination[T any] struct {
	Data         []T   `json:"data"`
	Page         int   `json:"page"`
	Total        int64 `json:"total"`
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
}

func Paginate[T any](data []T, page, pageSize int, total int64) *Pagination[T] {
	pSize := float64(pageSize)
	if pSize <= 0 {
		pSize = 10
	}

	totalPages := int(math.Ceil(float64(total) / pSize))

	pagination := Pagination[T]{
		Data:         data,
		Page:         page,
		Total:        total,
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
		NextPage:     nil,
		PreviousPage: nil,
	}

	var nextPage *int
	if pagination.Page < pagination.TotalPages {
		p := pagination.Page + 1
		nextPage = &p
	}

	var prevPage *int
	if pagination.Page > 1 && pagination.Page <= pagination.TotalPages {
		p := pagination.Page - 1
		prevPage = &p
	}

	pagination.NextPage = nextPage
	pagination.PreviousPage = prevPage

	return &pagination
}

// MapPaginatedResult transforms a paginated result containing items of a source type (S)
// into a new result containing items of a destination type (D).
//
// It takes a source Pagination and a mapper function that defines the conversion
// logic from an item of type S to an item of type D.
//
// Type Parameters:
//   - S: The source type (e.g., a database model like database.Post).
//   - D: The destination type (e.g., an API response DTO like PostResponse).
//
// The function returns a new Pagination with the transformed data, while preserving
// all original pagination metadata (Total, CurrentPage, etc.).
func MapPaginatedResult[S any, D any](source *Pagination[S], mapper func(S) D) *Pagination[D] {
	mappedData := make([]D, len(source.Data))

	// Iterate over the source data and apply the mapper function
	for i, item := range source.Data {
		mappedData[i] = mapper(item)
	}

	return &Pagination[D]{
		Data:         mappedData,
		Total:        source.Total,
		CurrentPage:  source.CurrentPage,
		PageSize:     source.PageSize,
		TotalPages:   source.TotalPages,
		NextPage:     source.NextPage,
		PreviousPage: source.PreviousPage,
	}
}
