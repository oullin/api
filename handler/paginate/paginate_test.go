package paginate_test

import (
	"net/url"
	"testing"

	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/paginate"
)

func TestNewFrom(t *testing.T) {
	u, _ := url.Parse("https://example.com/posts?page=2&limit=50")
	p := paginate.NewFrom(u, 5)

	if p.Page != 2 {
		t.Fatalf("expected page to be 2, got %d", p.Page)
	}

	if p.Limit != pagination.PostsMaxLimit {
		t.Fatalf("expected limit to be %d, got %d", pagination.PostsMaxLimit, p.Limit)
	}

	u2, _ := url.Parse("/categories?page=-1&limit=50")
	p2 := paginate.NewFrom(u2, 5)

	if p2.Page != pagination.MinPage || p2.Limit != pagination.CategoriesMaxLimit {
		t.Fatalf("expected page to be %d and limit to be %d, got %+v", pagination.MinPage, pagination.CategoriesMaxLimit, p2)
	}
}
