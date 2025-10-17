package payload

import (
	"testing"

	"github.com/oullin/database"
)

func TestGetCategoriesResponse(t *testing.T) {
	sort := 2
	cats := []database.Category{
		{
			UUID:        "1",
			Name:        "n",
			Slug:        "s",
			Description: "d",
			Sort:        &sort,
		},
	}

	r := GetCategoriesResponse(cats)

	if len(r) != 1 || r[0].Slug != "s" || r[0].Sort == nil || *r[0].Sort != 2 {
		t.Fatalf("unexpected %#v", r)
	}
}
