package images

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

const avifFixtureBase64 = "AAAAHGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZgAAAOptZXRhAAAAAAAAACFoZGxyAAAAAAAAAABwaWN0AAAAAAAAAAAAAAAAAAAAAA5waXRtAAAAAAABAAAAImlsb2MAAAAAREAAAQABAAAAAAEOAAEAAAAAAAAAHwAAACNpaW5mAAAAAAABAAAAFWluZmUCAAAAAAEAAGF2MDEAAAAAamlwcnAAAABLaXBjbwAAABNjb2xybmNseAABAA0ABoAAAAAMYXYxQ4EADAAAAAAUaXNwZQAAAAAAAAACAAAAAgAAABBwaXhpAAAAAAMICAgAAAAXaXBtYQAAAAAAAAABAAEEAYIDBAAAACdtZGF0EgAKCBgANggIaDQgMhEYAAooooQAALATVl9ApOsM/A=="

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
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/webp")
		if _, err := w.Write(data); err != nil {
			t.Errorf("write webp payload: %v", err)
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

func TestFetchRemoteJPEGWithLeadingNoise(t *testing.T) {
	t.Parallel()

	img := createTestImage(24, 16)
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}

	junkPrefix := []byte{0xEF, 0xBB, 0xBF, '\n', '\r', ' '}
	payload := append(junkPrefix, jpegBuf.Bytes()...)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.jpg" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write jpeg payload: %v", err)
		}
	}))
	defer server.Close()

	imgResult, format, err := Fetch(server.URL + "/cover.jpg")
	if err != nil {
		t.Fatalf("fetch jpeg with leading noise: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected jpeg format, got %q", format)
	}

	bounds := imgResult.Bounds()
	if bounds.Dx() != 24 || bounds.Dy() != 16 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteJPEGEmbeddedAfterHTML(t *testing.T) {
	t.Parallel()

	img := createTestImage(30, 22)
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}

	payload := append([]byte("<html><body>preview</body></html>"), jpegBuf.Bytes()...)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/asset" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write jpeg payload: %v", err)
		}
	}))
	defer server.Close()

	imgResult, format, err := Fetch(server.URL + "/asset")
	if err != nil {
		t.Fatalf("fetch embedded jpeg: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected jpeg format, got %q", format)
	}

	bounds := imgResult.Bounds()
	if bounds.Dx() != 30 || bounds.Dy() != 22 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteJPEGCompressedWithoutEncodingHeader(t *testing.T) {
	t.Parallel()

	img := createTestImage(40, 28)
	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}

	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	if _, err := gzipWriter.Write(jpegBuf.Bytes()); err != nil {
		t.Fatalf("compress jpeg: %v", err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.jpg" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(compressed.Bytes()); err != nil {
			t.Fatalf("write compressed payload: %v", err)
		}
	}))
	defer server.Close()

	imgResult, format, err := Fetch(server.URL + "/cover.jpg")
	if err != nil {
		t.Fatalf("fetch compressed jpeg: %v", err)
	}

	if format != "jpeg" {
		t.Fatalf("expected jpeg format, got %q", format)
	}

	bounds := imgResult.Bounds()
	if bounds.Dx() != 40 || bounds.Dy() != 28 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteSetsAcceptHeader(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.png" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "image/avif") {
			t.Errorf("unexpected avif accept header: %s", accept)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		if accept != supportedImageAcceptHeader {
			t.Errorf("unexpected accept header: %s", accept)
		}

		if ua := r.Header.Get("User-Agent"); ua != defaultRemoteImageUA {
			t.Errorf("unexpected user agent: %s", ua)
		}

		if err := png.Encode(w, createTestImage(10, 10)); err != nil {
			t.Errorf("encode png: %v", err)
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
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
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

func TestFetchRemoteBrotliWithoutEncodingHeader(t *testing.T) {
	t.Parallel()

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, createTestImage(28, 16)); err != nil {
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.png" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "image/png")
		if _, err := w.Write(brBuf.Bytes()); err != nil {
			t.Fatalf("write brotli payload: %v", err)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/cover.png")
	if err != nil {
		t.Fatalf("fetch brotli payload without header: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 28 || bounds.Dy() != 16 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteZstdEncoded(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cover.png" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, createTestImage(18, 12)); err != nil {
			t.Fatalf("encode png: %v", err)
		}

		var zstdBuf bytes.Buffer
		writer, err := zstd.NewWriter(&zstdBuf)
		if err != nil {
			t.Fatalf("create zstd writer: %v", err)
		}

		if _, err := writer.Write(pngBuf.Bytes()); err != nil {
			t.Fatalf("write zstd payload: %v", err)
		}

		if err := writer.Close(); err != nil {
			t.Fatalf("close zstd writer: %v", err)
		}

		w.Header().Set("Content-Encoding", "zstd")
		w.Header().Set("Content-Type", "image/png")
		if _, err := w.Write(zstdBuf.Bytes()); err != nil {
			t.Fatalf("write zstd payload: %v", err)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/cover.png")
	if err != nil {
		t.Fatalf("fetch zstd image: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 18 || bounds.Dy() != 12 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestFetchRemoteAVIFDecode(t *testing.T) {
	t.Parallel()

	avifData, err := base64.StdEncoding.DecodeString(avifFixtureBase64)
	if err != nil {
		t.Fatalf("decode avif fixture: %v", err)
	}

	var requests []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.String())

		if r.URL.Path != "/user-attachments/assets/e5abb532-59bf-49bb-a9d2-0c31872718d7" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(avifData); err != nil {
			t.Errorf("write avif payload: %v", err)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/user-attachments/assets/e5abb532-59bf-49bb-a9d2-0c31872718d7")
	if err != nil {
		t.Fatalf("fetch avif image: %v", err)
	}

	if format != "avif" {
		t.Fatalf("expected avif format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("unexpected avif dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	if len(requests) != 1 {
		t.Fatalf("expected single request, got %v", requests)
	}
}

func TestFetchRemoteAVIFGitHubFallback(t *testing.T) {
	t.Parallel()

	avifData, err := base64.StdEncoding.DecodeString(avifFixtureBase64)
	if err != nil {
		t.Fatalf("decode avif fixture: %v", err)
	}

	brokenAvif := append([]byte(nil), avifData...)
	if len(brokenAvif) > 40 {
		brokenAvif = brokenAvif[:40]
	}

	pngImage := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	pngImage.Set(0, 0, color.NRGBA{R: 200, G: 10, B: 10, A: 255})
	pngImage.Set(1, 1, color.NRGBA{R: 20, G: 200, B: 20, A: 255})

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, pngImage); err != nil {
		t.Fatalf("encode fallback png: %v", err)
	}

	const attachmentID = "e5abb532-59bf-49bb-a9d2-0c31872718d7"

	type requestInfo struct {
		URL    string
		Accept string
	}

	var requests []requestInfo

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, requestInfo{URL: r.URL.String(), Accept: r.Header.Get("Accept")})

		switch {
		case r.URL.Path == "/user-attachments/assets/"+attachmentID && r.URL.RawQuery == "":
			w.Header().Set("Content-Type", "image/jpeg")
			if _, err := w.Write(brokenAvif); err != nil {
				t.Errorf("write avif payload: %v", err)
			}
		case r.URL.Path == "/user-attachments/assets/"+attachmentID && r.URL.Query().Get("format") == "png":
			w.Header().Set("Content-Type", "image/png")
			if _, err := w.Write(pngBuf.Bytes()); err != nil {
				t.Errorf("write png payload: %v", err)
			}
		default:
			t.Errorf("unexpected request: %s", r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	img, format, err := Fetch(server.URL + "/user-attachments/assets/" + attachmentID)
	if err != nil {
		t.Fatalf("fetch remote avif fallback: %v", err)
	}

	if format != "png" {
		t.Fatalf("expected png format, got %q", format)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Fatalf("unexpected fallback dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}

	if len(requests) < 2 {
		t.Fatalf("expected fallback requests, got %v", requests)
	}

	if got := requests[0].Accept; got != supportedImageAcceptHeader {
		t.Fatalf("unexpected accept header for initial request: %s", got)
	}

	var sawPNG bool
	for _, req := range requests[1:] {
		if strings.Contains(req.URL, "format=png") {
			if req.Accept != fallbackPNGAcceptHeader {
				t.Fatalf("expected png fallback accept header %q, got %q", fallbackPNGAcceptHeader, req.Accept)
			}
			sawPNG = true
			break
		}
	}

	if !sawPNG {
		t.Fatalf("expected png fallback request, got %v", requests)
	}
}

func TestFetchRemoteDecodeErrorIncludesContentType(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html></html>"))
	}))
	defer server.Close()

	_, _, err := Fetch(server.URL + "/cover.png")
	if err == nil {
		t.Fatal("expected error decoding html response")
	}

	if !strings.Contains(err.Error(), "content-type \"text/html\"") {
		t.Fatalf("expected content-type in error, got %v", err)
	}
}

func TestFetchRemoteDecodeErrorIncludesContentEncoding(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "br")
		w.Header().Set("Content-Type", "text/html")

		var brBuf bytes.Buffer
		writer := brotli.NewWriterLevel(&brBuf, brotli.BestCompression)
		if _, err := writer.Write([]byte("<html></html>")); err != nil {
			t.Fatalf("write brotli payload: %v", err)
		}

		if err := writer.Close(); err != nil {
			t.Fatalf("close brotli writer: %v", err)
		}

		if _, err := w.Write(brBuf.Bytes()); err != nil {
			t.Fatalf("write brotli payload: %v", err)
		}
	}))
	defer server.Close()

	_, _, err := Fetch(server.URL + "/cover.png")
	if err == nil {
		t.Fatal("expected error decoding brotli encoded html")
	}

	if !strings.Contains(err.Error(), "content-encoding \"br\"") {
		t.Fatalf("expected content-encoding in error, got %v", err)
	}

	if !strings.Contains(err.Error(), "content-type \"text/html\"") {
		t.Fatalf("expected content-type in error, got %v", err)
	}
}

func TestFetchRemoteDecodeErrorProvidesDiagnostics(t *testing.T) {
	t.Parallel()

	payload := []byte("<html>forbidden</html>")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write payload: %v", err)
		}
	}))
	defer server.Close()

	_, _, err := Fetch(server.URL + "/cover.jpg")
	if err == nil {
		t.Fatal("expected decode error")
	}

	var decodeErr *DecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected DecodeError, got %T", err)
	}

	if decodeErr.SniffedType == "" || !strings.Contains(decodeErr.SniffedType, "text/html") {
		t.Fatalf("expected sniffed type to mention text/html, got %q", decodeErr.SniffedType)
	}

	if decodeErr.Size != len(payload) {
		t.Fatalf("unexpected payload size: got %d want %d", decodeErr.Size, len(payload))
	}

	if decodeErr.Hash == "" {
		t.Fatal("expected hash to be populated")
	}

	if !strings.HasPrefix(decodeErr.PrefixHex, "3c68") { // "<h"
		t.Fatalf("expected prefix to include html bytes, got %s", decodeErr.PrefixHex)
	}

	diagnostics := decodeErr.Diagnostics()
	var sawPrefix bool
	for _, line := range diagnostics {
		if strings.HasPrefix(line, "payload-prefix-hex:") {
			sawPrefix = true
		}
	}

	if !sawPrefix {
		t.Fatalf("expected diagnostics to include payload prefix, got %v", diagnostics)
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

	if !webpEncodeSupported() {
		t.Skip("webp encoding requires cgo")
	}

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
