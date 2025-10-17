package payload

import (
	"testing"

	"github.com/oullin/database"
)

func TestGetCategoriesResponse(t *testing.T) {
	cats := []database.Category{
		{
			UUID:        "1",
			Name:        "n",
			Slug:        "s",
			Description: "d",
			Sort:        2,
		},
	}

	r := GetCategoriesResponse(cats)

	if len(r) != 1 || r[0].Slug != "s" || r[0].Sort != 2 {
		t.Fatalf("unexpected %#v", r)
	}
}
