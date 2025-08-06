package payload

import (
	"testing"

	"github.com/oullin/database"
)

func TestGetTagsResponse(t *testing.T) {
	tags := []database.Tag{{UUID: "1", Name: "n", Slug: "s", Description: "d"}}
	r := GetTagsResponse(tags)
	if len(r) != 1 || r[0].Slug != "s" {
		t.Fatalf("unexpected %#v", r)
	}
}
