package image

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
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

const DefaultJPEGQuality = 85

func Fetch(source string) (stdimage.Image, string, error) {
	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", fmt.Errorf("parse url: %w", err)
	}

	reader, err := openSource(parsed)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()

	img, format, err := stdimage.Decode(reader)
	if err != nil {
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
	ext := strings.ToLower(filepath.Ext(source))
	if ext == "" {
		ext = "." + strings.ToLower(format)
	}

	switch ext {
	case ".jpeg":
		return ".jpg"
	case ".jpg", ".png":
		return ext
	default:
		if format == "png" {
			return ".png"
		}

		return ".jpg"
	}
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
	default:
		return "image/png"
	}
}

func openSource(parsed *url.URL) (io.ReadCloser, error) {
	switch parsed.Scheme {
	case "http", "https":
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(parsed.String())
		if err != nil {
			return nil, fmt.Errorf("download image: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer resp.Body.Close()
			return nil, fmt.Errorf("download image: unexpected status %s", resp.Status)
		}

		return resp.Body, nil
	case "file":
		return openLocal(parsed)
	case "":
		return os.Open(parsed.Path)
	default:
		return nil, fmt.Errorf("unsupported image scheme: %s", parsed.Scheme)
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
