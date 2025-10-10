package images

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	_ "image/jpeg"

	"github.com/andybalholm/brotli"
)

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 200, A: 255})
		}
	}

	return img
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()

	fh, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	defer fh.Close()

	if err := png.Encode(fh, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
}

func TestFetchLocal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "local.png")
	writePNG(t, src, createTestImage(100, 80))

	fileURL := url.URL{Scheme: "file", Path: src}

	img, format, err := Fetch(fileURL.String())
	if err != nil {
		t.Fatalf("fetch local image: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 80 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemote(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.png" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		if err := png.Encode(w, createTestImage(50, 40)); err != nil {
			t.Fatalf("encode remote png: %v", err)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/cover.png")
	if err != nil {
		t.Fatalf("fetch remote image: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 40 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteWebP(t *testing.T) {
	t.Parallel()

	const webpBase64 = "UklGRjwAAABXRUJQVlA4IDAAAADQAQCdASoBAAEAAUAmJaACdLoB+AADsAD+8ut//NgVzXPv9//S4P0uD9Lg/9KQAAA="

	data, err := base64.StdEncoding.DecodeString(webpBase64)
	if err != nil {
		t.Fatalf("decode webp fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.webp" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "image/webp")
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write webp payload: %v", err)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/cover.webp")
	if err != nil {
		t.Fatalf("fetch remote webp: %v", err)
	}

	if format != "webp" {
		t.Fatalf("expected webp format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 1 || bounds.Dy() != 1 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteBrotliEncoded(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.png" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, createTestImage(32, 24)); err != nil {
			t.Fatalf("encode png: %v", err)
		}

		var brBuf bytes.Buffer
		writer := brotli.NewWriterLevel(&brBuf, brotli.BestCompression)
		if _, err := writer.Write(pngBuf.Bytes()); err != nil {
			t.Fatalf("write brotli: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("close brotli writer: %v", err)
		}

		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Content-Type", "image/png")
		if _, err := w.Write(brBuf.Bytes()); err != nil {
			t.Fatalf("write brotli payload: %v", err)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/cover.png")
	if err != nil {
		t.Fatalf("fetch brotli image: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 24 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResize(t *testing.T) {
	t.Parallel()

	src := createTestImage(20, 10)
	resized := Resize(src, 200, 100)

	bounds := resized.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestDetermineExtension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		format string
		want   string
	}{
		{"explicit png", "example.png", "jpeg", ".png"},
		{"explicit jpg", "example.jpg", "png", ".jpg"},
		{"explicit jpeg", "example.jpeg", "png", ".jpg"},
		{"explicit webp", "example.webp", "jpeg", ".webp"},
		{"missing ext png format", "example", "png", ".png"},
		{"missing ext jpeg format", "example", "jpeg", ".jpg"},
		{"missing ext webp format", "example", "webp", ".webp"},
		{"unknown ext falls back to jpg", "example.gif", "jpeg", ".jpg"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := DetermineExtension(tc.source, tc.format); got != tc.want {
				t.Fatalf("DetermineExtension(%q, %q) = %q, want %q", tc.source, tc.format, got, tc.want)
			}
		})
	}
}

func TestBuildFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		slug     string
		ext      string
		fallback string
		want     string
	}{
		{"uses slug", "my-post", ".png", "fallback", "my-post.png"},
		{"trims whitespace", "  my post  ", ".jpg", "fallback", "my-post.jpg"},
		{"removes leading slash", "/post/slug/", ".png", "fallback", "post/slug.png"},
		{"uses fallback", "   ", ".jpg", "fallback", "fallback.jpg"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := BuildFileName(tc.slug, tc.ext, tc.fallback); got != tc.want {
				t.Fatalf("BuildFileName(%q, %q, %q) = %q, want %q", tc.slug, tc.ext, tc.fallback, got, tc.want)
			}
		})
	}
}

func TestMIMEFromExtension(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		".png":  "image/png",
		".PNG":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/png",
		".gif":  "image/png",
		".webp": "image/webp",
	}

	for ext, want := range tests {
		ext := ext
		want := want
		t.Run(ext, func(t *testing.T) {
			t.Parallel()

			if got := MIMEFromExtension(ext); got != want {
				t.Fatalf("MIMEFromExtension(%q) = %q, want %q", ext, got, want)
			}
		})
	}
}

func TestSaveAndMove(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "source.png")
	dst := filepath.Join(dir, "dest.png")

	// Create an existing destination to ensure Move removes it first.
	writePNG(t, dst, createTestImage(10, 10))

	if err := Save(src, createTestImage(60, 30), ".png", DefaultJPEGQuality); err != nil {
		t.Fatalf("save png: %v", err)
	}

	if err := Move(src, dst); err != nil {
		t.Fatalf("move file: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("expected source to be removed, got %v", err)
	}

	fh, err := os.Open(dst)
	if err != nil {
		t.Fatalf("open dest: %v", err)
	}
	defer fh.Close()

	img, format, err := image.Decode(fh)
	if err != nil {
		t.Fatalf("decode moved image: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 60 || bounds.Dy() != 30 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestSaveJPEG(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "photo.jpg")

	if err := Save(path, createTestImage(25, 25), ".jpg", 70); err != nil {
		t.Fatalf("save jpg: %v", err)
	}

	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("open jpeg: %v", err)
	}
	defer fh.Close()

	img, format, err := image.Decode(fh)
	if err != nil {
		t.Fatalf("decode jpeg: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected jpeg format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 25 || bounds.Dy() != 25 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestSaveWebP(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "photo.webp")

	if err := Save(path, createTestImage(40, 20), ".webp", 80); err != nil {
		t.Fatalf("save webp: %v", err)
	}

	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("open webp: %v", err)
	}
	defer fh.Close()

	img, format, err := image.Decode(fh)
	if err != nil {
		t.Fatalf("decode webp: %v", err)
	}

	if format != "webp" {
		t.Fatalf("expected webp format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 40 || bounds.Dy() != 20 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestNormalizeRelativeURL(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  string
	}{
		"empty string": {
			input: "",
			want:  "",
		},
		"single dot": {
			input: ".",
			want:  "",
		},
		"root slash": {
			input: "/",
			want:  "",
		},
		"double dot": {
			input: "..",
			want:  "",
		},
		"triple dot": {
			input: "...",
			want:  "...",
		},
		"nested traversal": {
			input: "../../foo/bar.png",
			want:  "foo/bar.png",
		},
		"leading traversal": {
			input: "../foo/bar.png",
			want:  "foo/bar.png",
		},
		"leading slash": {
			input: "/foo/bar.png",
			want:  "foo/bar.png",
		},
		"current dir segments": {
			input: "./foo/./bar.png",
			want:  "foo/bar.png",
		},
		"current dir prefix": {
			input: "./foo.png",
			want:  "foo.png",
		},
		"cleanup mixed": {
			input: "foo/../bar/baz.png",
			want:  "bar/baz.png",
		},
		"trailing slash": {
			input: "foo/bar/",
			want:  "foo/bar",
		},
		"windows separators": {
			input: "..\\foo\\bar\\baz.png",
			want:  "foo/bar/baz.png",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := NormalizeRelativeURL(tc.input); got != tc.want {
				t.Fatalf("NormalizeRelativeURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
