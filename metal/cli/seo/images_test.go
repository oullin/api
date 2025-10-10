package seo

import (
	"image"
	"image/color"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
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

func TestGeneratorPreparePostImageNormalizesRelativeURL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	spaDir := filepath.Join(root, "seo")
	imagesDir := filepath.Join(root, "images", "seo")

	if err := os.MkdirAll(spaDir, 0o755); err != nil {
		t.Fatalf("create spa dir: %v", err)
	}

	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		t.Fatalf("create images dir: %v", err)
	}

	srcPath := filepath.Join(root, "source.png")
	fh, err := os.Create(srcPath)
	if err != nil {
		t.Fatalf("create source image: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}

	if err := png.Encode(fh, img); err != nil {
		t.Fatalf("encode source image: %v", err)
	}

	if err := fh.Close(); err != nil {
		t.Fatalf("close source image: %v", err)
	}

	fileURL := url.URL{Scheme: "file", Path: srcPath}

	gen := &Generator{
		Page: Page{
			SiteName:  "SEO Test Suite",
			SiteURL:   "https://seo.example.test",
			OutputDir: spaDir,
		},
		Env: &env.Environment{Seo: env.SeoEnvironment{SpaDir: spaDir, SpaImagesDir: imagesDir}},
	}

	post := payload.PostResponse{Slug: "awesome-post", CoverImageURL: fileURL.String()}

	prepared, err := gen.preparePostImage(post)
	if err != nil {
		t.Fatalf("prepare post image: %v", err)
	}

	if strings.Contains(prepared.URL, "../") {
		t.Fatalf("expected url without parent traversal: %s", prepared.URL)
	}

	expectedPrefix := "https://seo.example.test/images/seo/"
	if !strings.HasPrefix(prepared.URL, expectedPrefix) {
		t.Fatalf("unexpected url prefix: got %s, want prefix %s", prepared.URL, expectedPrefix)
	}

	destPath := filepath.Join(imagesDir, "awesome-post.png")
	if _, err := os.Stat(destPath); err != nil {
		t.Fatalf("stat destination image: %v", err)
	}
}
