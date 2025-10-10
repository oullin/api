package images

import (
	"bytes"
	"compress/gzip"
	"errors"
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
	"unicode"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/andybalholm/brotli"
	"github.com/chai2010/webp"
	"github.com/klauspost/compress/zstd"
)

func Fetch(source string) (stdimage.Image, string, error) {
	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", fmt.Errorf("parse url: %w", err)
	}

	reader, contentType, encoding, err := openSource(parsed)
	if err != nil {
		return nil, "", err
	}
	payload, err := readImagePayload(reader)
	if err != nil {
		return nil, "", err
	}

	img, format, err := decodeImagePayload(payload)
	if err != nil {
		var details []string

		if ct := strings.TrimSpace(contentType); ct != "" {
			details = append(details, fmt.Sprintf("content-type %q", ct))
		}

		if enc := strings.TrimSpace(encoding); enc != "" {
			details = append(details, fmt.Sprintf("content-encoding %q", enc))
		}

		if len(details) > 0 {
			return nil, "", fmt.Errorf("decode image (%s): %w", strings.Join(details, ", "), err)
		}

		return nil, "", fmt.Errorf("decode image: %w", err)
	}

	return img, format, nil
}

const maxRemoteImageBytes = 32 << 20 // 32MiB should cover large blog assets.

func readImagePayload(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()

	limited := io.LimitReader(reader, maxRemoteImageBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read image payload: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("empty image payload")
	}

	if len(data) > maxRemoteImageBytes {
		return nil, fmt.Errorf("image payload exceeds %d bytes", maxRemoteImageBytes)
	}

	return data, nil
}

func decodeImagePayload(data []byte) (stdimage.Image, string, error) {
	attempts := [][]byte{data}

	trimmed := trimLeadingNoise(data)
	if len(trimmed) > 0 && !bytes.Equal(trimmed, data) {
		attempts = append(attempts, trimmed)
	}

	if start, ok := findEmbeddedImageStart(trimmed); ok && start > 0 && start < len(trimmed) {
		attempts = append(attempts, trimmed[start:])
	}

	var lastErr error
	for _, candidate := range attempts {
		img, format, err := stdimage.Decode(bytes.NewReader(candidate))
		if err == nil {
			return img, format, nil
		}

		lastErr = err
	}

	if lastErr == nil {
		lastErr = errors.New("image: unknown format")
	}

	return nil, "", lastErr
}

func trimLeadingNoise(data []byte) []byte {
	trimmed := bytes.TrimLeftFunc(data, func(r rune) bool {
		return r == unicode.ReplacementChar || unicode.IsSpace(r)
	})

	if len(trimmed) >= 3 && bytes.Equal(trimmed[:3], []byte{0xEF, 0xBB, 0xBF}) {
		trimmed = trimmed[3:]
	}

	return trimmed
}

func findEmbeddedImageStart(data []byte) (int, bool) {
	if idx := bytes.Index(data, []byte{0xFF, 0xD8, 0xFF}); idx >= 0 {
		return idx, true
	}

	if idx := bytes.Index(data, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}); idx >= 0 {
		return idx, true
	}

	if idx := bytes.Index(data, []byte("GIF87a")); idx >= 0 {
		return idx, true
	}

	if idx := bytes.Index(data, []byte("GIF89a")); idx >= 0 {
		return idx, true
	}

	for idx := bytes.Index(data, []byte("RIFF")); idx >= 0; {
		if len(data)-idx >= 12 && bytes.Equal(data[idx+8:idx+12], []byte("WEBP")) {
			return idx, true
		}

		next := bytes.Index(data[idx+4:], []byte("RIFF"))
		if next < 0 {
			break
		}

		idx += 4 + next
	}

	return 0, false
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

func openSource(parsed *url.URL) (io.ReadCloser, string, string, error) {
	switch parsed.Scheme {
	case "http", "https":
		client := &http.Client{Timeout: 10 * time.Second}

		req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, "", "", fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Accept", supportedImageAcceptHeader)
		req.Header.Set("User-Agent", defaultRemoteImageUA)

		resp, err := client.Do(req)
		if err != nil {
			return nil, "", "", fmt.Errorf("download image: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer resp.Body.Close()
			return nil, "", "", fmt.Errorf("download image: unexpected status %s", resp.Status)
		}

		reader, encoding, err := wrapHTTPBody(resp)
		if err != nil {
			return nil, "", "", err
		}

		return reader, resp.Header.Get("Content-Type"), encoding, nil
	case "file":
		reader, err := openLocal(parsed)
		return reader, "", "", err
	case "":
		reader, err := os.Open(parsed.Path)
		return reader, "", "", err
	default:
		return nil, "", "", fmt.Errorf("unsupported image scheme: %s", parsed.Scheme)
	}
}

func wrapHTTPBody(resp *http.Response) (io.ReadCloser, string, error) {
	encoding := strings.TrimSpace(strings.ToLower(resp.Header.Get("Content-Encoding")))
	if idx := strings.IndexRune(encoding, ','); idx >= 0 {
		encoding = encoding[:idx]
	}
	switch encoding {
	case "", "identity":
		return resp.Body, encoding, nil
	case "br":
		return composedReadCloser{Reader: brotli.NewReader(resp.Body), Closer: resp.Body}, encoding, nil
	case "gzip":
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, encoding, fmt.Errorf("prepare gzip decoder: %w", err)
		}

		return composedReadCloser{Reader: reader, Closer: multiCloser{reader, resp.Body}}, encoding, nil
	case "zstd", "zstandard":
		decoder, err := zstd.NewReader(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, encoding, fmt.Errorf("prepare zstd decoder: %w", err)
		}

		return composedReadCloser{Reader: decoder, Closer: multiCloser{noErrorCloseFunc(decoder.Close), resp.Body}}, encoding, nil
	default:
		return resp.Body, encoding, nil
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
