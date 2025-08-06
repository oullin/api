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
		},
	}
	r := GetCategoriesResponse(cats)
	if len(r) != 1 || r[0].Slug != "s" {
		t.Fatalf("unexpected %#v", r)
	}
}
