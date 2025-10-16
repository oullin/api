package paginate

import (
	"net/url"
	"testing"

	"github.com/oullin/database/repository/pagination"
)

func TestNewFrom(t *testing.T) {
	u, _ := url.Parse("https://example.com/posts?page=2&limit=50")
	p := NewFrom(u, 5)

	if p.Page != 2 {
		t.Fatalf("page %d", p.Page)
	}

	if p.Limit != pagination.PostsMaxLimit {
		t.Fatalf("limit %d", p.Limit)
	}

	u2, _ := url.Parse("/categories?page=-1&limit=50")
	p2 := NewFrom(u2, 5)

	if p2.Page != pagination.MinPage || p2.Limit != pagination.CategoriesMaxLimit {
		t.Fatalf("unexpected %+v", p2)
	}
}
