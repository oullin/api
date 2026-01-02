package pagination_test

import (
	"testing"

	"github.com/oullin/database/repository/pagination"
)

func TestNewPagination(t *testing.T) {
	p := pagination.Paginate{
		Page:  2,
		Limit: 2,
	}

	p.SetNumItems(5)

	result := pagination.NewPagination([]int{1, 2}, p)

	if result.TotalPages != 3 {
		t.Fatalf("expected 3 pages, got %d", result.TotalPages)
	}

	if result.NextPage == nil || *result.NextPage != 3 {
		t.Fatalf("expected next page to be 3, got mismatch")
	}

	if result.PreviousPage == nil || *result.PreviousPage != 1 {
		t.Fatalf("expected previous page to be 1, got mismatch")
	}
}

func TestHydratePagination(t *testing.T) {
	src := &pagination.Pagination[string]{
		Data:       []string{"a", "bb"},
		Page:       1,
		Total:      2,
		PageSize:   2,
		TotalPages: 1,
	}

	dst := pagination.HydratePagination(src, func(s string) int { return len(s) })

	if len(dst.Data) != 2 || dst.Data[1] != 2 {
		t.Fatalf("expected hydrated data with 2 items where second item is 2, got unexpected result")
	}

	if dst.Total != src.Total || dst.Page != src.Page {
		t.Fatalf("expected metadata to match source, got mismatch")
	}
}
