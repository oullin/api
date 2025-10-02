package seo

import (
	"errors"
	"fmt"
	"image"
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

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/portal"
	"golang.org/x/image/draw"
)

type preparedImage struct {
	URL  string
	Mime string
}

const (
	seoStorageDir       = "storage/seo"
	postImagesDir       = "posts"
	postImagesFolder    = "images"
	seoImageWidth       = 1200
	seoImageHeight      = 630
	defaultImageQuality = 85
)

func (g *Generator) preparePostImage(post payload.PostResponse) (preparedImage, error) {
	source := strings.TrimSpace(post.CoverImageURL)
	if source == "" {
		return preparedImage{}, errors.New("post has no cover image url")
	}

	img, format, err := fetchImage(source)
	if err != nil {
		return preparedImage{}, err
	}

	resized := resizeImage(img)

	ext := determineExtension(source, format)
	fileName := buildImageFileName(post.Slug, ext)

	if err := os.MkdirAll(seoStorageDir, 0o755); err != nil {
		return preparedImage{}, fmt.Errorf("create storage dir: %w", err)
	}

	tempPath := filepath.Join(seoStorageDir, fileName)
	if err := saveImage(tempPath, resized, ext); err != nil {
		return preparedImage{}, fmt.Errorf("write resized image: %w", err)
	}

	defer func() {
		_ = os.Remove(tempPath)
	}()

	destDir := filepath.Join(g.Page.OutputDir, postImagesDir, postImagesFolder)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return preparedImage{}, fmt.Errorf("create destination dir: %w", err)
	}

	destPath := filepath.Join(destDir, fileName)
	if err := moveFile(tempPath, destPath); err != nil {
		return preparedImage{}, fmt.Errorf("move resized image: %w", err)
	}

	relative := path.Join(postImagesDir, postImagesFolder, fileName)

	return preparedImage{
		URL:  g.siteURLFor(relative),
		Mime: mimeFromExtension(ext),
	}, nil
}

func fetchImage(source string) (image.Image, string, error) {
	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", fmt.Errorf("parse url: %w", err)
	}

	reader, err := openImageSource(parsed)
	if err != nil {
		return nil, "", err
	}
	defer reader.Close()

	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("decode image: %w", err)
	}

	return img, format, nil
}

func openImageSource(parsed *url.URL) (io.ReadCloser, error) {
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
		return openLocalFile(parsed)
	case "":
		// Treat empty scheme as local file path
		return os.Open(parsed.Path)
	default:
		return nil, fmt.Errorf("unsupported image scheme: %s", parsed.Scheme)
	}
}

func openLocalFile(parsed *url.URL) (io.ReadCloser, error) {
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

func resizeImage(img image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, seoImageWidth, seoImageHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	return dst
}

func determineExtension(source, format string) string {
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

func buildImageFileName(slug, ext string) string {
	trimmed := strings.TrimSpace(slug)
	cleaned := strings.Trim(trimmed, "/")
	if cleaned == "" {
		cleaned = "post-image"
	}

	cleaned = strings.ReplaceAll(cleaned, " ", "-")

	return cleaned + ext
}

func saveImage(path string, img image.Image, ext string) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	switch ext {
	case ".png":
		return pngEncode(fh, img)
	default:
		return jpegEncode(fh, img)
	}
}

func pngEncode(w io.Writer, img image.Image) error {
	encoder := &png.Encoder{CompressionLevel: png.DefaultCompression}
	return encoder.Encode(w, img)
}

func jpegEncode(w io.Writer, img image.Image) error {
	options := &jpeg.Options{Quality: defaultImageQuality}
	return jpeg.Encode(w, img, options)
}

func moveFile(src, dst string) error {
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

func (g *Generator) siteURLFor(rel string) string {
	base := strings.TrimSuffix(g.Page.SiteURL, "/")
	rel = strings.TrimPrefix(rel, "/")

	if base == "" {
		return rel
	}

	return portal.SanitiseURL(base + "/" + rel)
}

func mimeFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg":
		return "image/jpeg"
	default:
		return "image/png"
	}
}
