package pagination

import "testing"

func TestMakePagination(t *testing.T) {
	p := Paginate{Page: 2, Limit: 2}
	p.SetNumItems(5)

	result := MakePagination([]int{1, 2}, p)

	if result.TotalPages != 3 {
		t.Fatalf("expected 3 pages got %d", result.TotalPages)
	}
	if result.NextPage == nil || *result.NextPage != 3 {
		t.Fatalf("next page mismatch")
	}
	if result.PreviousPage == nil || *result.PreviousPage != 1 {
		t.Fatalf("prev page mismatch")
	}
}

func TestHydratePagination(t *testing.T) {
	src := &Pagination[string]{
		Data:       []string{"a", "bb"},
		Page:       1,
		Total:      2,
		PageSize:   2,
		TotalPages: 1,
	}

	dst := HydratePagination(src, func(s string) int { return len(s) })

	if len(dst.Data) != 2 || dst.Data[1] != 2 {
		t.Fatalf("unexpected hydration")
	}
	if dst.Total != src.Total || dst.Page != src.Page {
		t.Fatalf("metadata mismatch")
	}
}
