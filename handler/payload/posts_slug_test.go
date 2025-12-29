package payload_test

import (
	"net/http/httptest"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestGetSlugFrom(t *testing.T) {
	r := httptest.NewRequest("GET", "/posts/s", nil)
	r.SetPathValue("slug", "  SLUG ")

	if s := payload.GetSlugFrom(r); s != "slug" {
		t.Fatalf("slug %s", s)
	}
}
