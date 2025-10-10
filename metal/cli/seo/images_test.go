package seo

import (
	"strings"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestGeneratorPreparePostImageMissingURL(t *testing.T) {
	gen := &Generator{Page: Page{}}

	_, err := gen.preparePostImage(payload.PostResponse{Slug: "test-post"})
	if err == nil {
		t.Fatalf("expected error for missing cover image url")
	}

	if !strings.Contains(err.Error(), "post has no cover image url") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGeneratorSiteURLFor(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		base string
		rel  string
		want string
	}{
		"with base and leading slash": {
			base: "https://example.test/",
			rel:  "/posts/images/pic.png",
			want: "https://example.test/posts/images/pic.png",
		},
		"without trailing slash": {
			base: "https://example.test",
			rel:  "posts/images/pic.png",
			want: "https://example.test/posts/images/pic.png",
		},
		"no base url": {
			base: "",
			rel:  "/posts/images/pic.png",
			want: "posts/images/pic.png",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gen := &Generator{Page: Page{SiteURL: tc.base}}
			got := gen.siteURLFor(tc.rel)

			if got != tc.want {
				t.Fatalf("siteURLFor(%q, %q) = %q, want %q", tc.base, tc.rel, got, tc.want)
			}
		})
	}
}
