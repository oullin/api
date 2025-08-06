package payload

import (
	"net/http/httptest"
	"testing"
)

func TestGetSlugFrom(t *testing.T) {
	r := httptest.NewRequest("GET", "/posts/s", nil)
	r.SetPathValue("slug", "  SLUG ")
	if s := GetSlugFrom(r); s != "slug" {
		t.Fatalf("slug %s", s)
	}
}
