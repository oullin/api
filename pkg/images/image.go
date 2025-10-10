package images

import (
	"fmt"
	stdimage "image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/andybalholm/brotli"
	"github.com/chai2010/webp"
)

func Fetch(source string) (stdimage.Image, string, error) {
	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", fmt.Errorf("parse url: %w", err)
	}

	reader, contentType, err := openSource(parsed)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()

	img, format, err := stdimage.Decode(reader)
	if err != nil {
		ct := strings.TrimSpace(contentType)
		if ct != "" {
			return nil, "", fmt.Errorf("decode image (content-type %q): %w", ct, err)
		}
		return nil, "", fmt.Errorf("decode image: %w", err)
	}

	return img, format, nil
}

func Resize(src stdimage.Image, width, height int) stdimage.Image {
	dst := stdimage.NewRGBA(stdimage.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return dst
}

func DetermineExtension(source, format string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(source)))
	format = strings.ToLower(strings.TrimSpace(format))

	switch ext {
	case ".jpeg":
		return ".jpg"
	case ".jpg", ".png", ".webp":
		return ext
	}

	switch format {
	case "jpeg", "jpg":
		return ".jpg"
	case "png":
		return ".png"
	case "webp":
		return ".webp"
	}

	return ".jpg"
}

func BuildFileName(slug, ext, fallback string) string {
	trimmed := strings.TrimSpace(slug)
	cleaned := strings.Trim(trimmed, "/")
	if cleaned == "" {
		cleaned = fallback
	}

	cleaned = strings.ReplaceAll(cleaned, " ", "-")

	return cleaned + ext
}

func Save(path string, img stdimage.Image, ext string, quality int) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	switch ext {
	case ".png":
		encoder := &png.Encoder{CompressionLevel: png.DefaultCompression}
		return encoder.Encode(fh, img)
	case ".webp":
		options := &webp.Options{Lossless: false, Quality: float32(quality)}
		return webp.Encode(fh, img, options)
	default:
		options := &jpeg.Options{Quality: quality}
		return jpeg.Encode(fh, img, options)
	}
}

func Move(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	return os.Remove(src)
}

func MIMEFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func NormalizeRelativeURL(rel string) string {
	rel = strings.ReplaceAll(rel, "\\", "/")

	cleaned := path.Clean(rel)

	if cleaned == "." || cleaned == "/" {
		return ""
	}

	parts := strings.Split(cleaned, "/")

	var b strings.Builder

	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}

		if b.Len() > 0 {
			b.WriteByte('/')
		}

		b.WriteString(part)
	}

	return b.String()
}

func openSource(parsed *url.URL) (io.ReadCloser, string, error) {
	switch parsed.Scheme {
	case "http", "https":
		client := &http.Client{Timeout: 10 * time.Second}

		req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, "", fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Accept", supportedImageAcceptHeader)
		req.Header.Set("User-Agent", defaultRemoteImageUA)

		resp, err := client.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("download image: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer resp.Body.Close()
			return nil, "", fmt.Errorf("download image: unexpected status %s", resp.Status)
		}

		return wrapHTTPBody(resp), resp.Header.Get("Content-Type"), nil
	case "file":
		reader, err := openLocal(parsed)
		return reader, "", err
	case "":
		reader, err := os.Open(parsed.Path)
		return reader, "", err
	default:
		return nil, "", fmt.Errorf("unsupported image scheme: %s", parsed.Scheme)
	}
}

func wrapHTTPBody(resp *http.Response) io.ReadCloser {
	encoding := strings.TrimSpace(strings.ToLower(resp.Header.Get("Content-Encoding")))
	if idx := strings.IndexRune(encoding, ','); idx >= 0 {
		encoding = encoding[:idx]
	}
	switch encoding {
	case "", "identity":
		return resp.Body
	case "br":
		return composedReadCloser{Reader: brotli.NewReader(resp.Body), Closer: resp.Body}
	default:
		return resp.Body
	}
}

func openLocal(parsed *url.URL) (io.ReadCloser, error) {
	pathValue := parsed.Path
	if pathValue == "" {
		pathValue = parsed.Opaque
	}

	if parsed.Host != "" {
		pathValue = "//" + parsed.Host + pathValue
	}

	unescaped, err := url.PathUnescape(pathValue)
	if err != nil {
		return nil, fmt.Errorf("decode file path: %w", err)
	}

	return os.Open(unescaped)
}
