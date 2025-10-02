package seo

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/oullin/handler/payload"
	pkgimages "github.com/oullin/pkg/images"
	"github.com/oullin/pkg/portal"
)

type preparedImage struct {
	URL  string
	Mime string
}

const (
	seoStorageDir    = "storage/seo"
	postImagesDir    = "posts"
	postImagesFolder = "images"
	seoImageWidth    = 1200
	seoImageHeight   = 630
)

func (g *Generator) preparePostImage(post payload.PostResponse) (preparedImage, error) {
	source := strings.TrimSpace(post.CoverImageURL)
	if source == "" {
		return preparedImage{}, errors.New("post has no cover image url")
	}

	img, format, err := pkgimages.Fetch(source)
	if err != nil {
		return preparedImage{}, err
	}

	resized := pkgimages.Resize(img, seoImageWidth, seoImageHeight)

	ext := pkgimages.DetermineExtension(source, format)
	fileName := pkgimages.BuildFileName(post.Slug, ext, "post-image")

	if err := os.MkdirAll(seoStorageDir, 0o755); err != nil {
		return preparedImage{}, fmt.Errorf("create storage dir: %w", err)
	}

	tempPath := filepath.Join(seoStorageDir, fileName)
	if err := pkgimages.Save(tempPath, resized, ext, pkgimages.DefaultJPEGQuality); err != nil {
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
	if err := pkgimages.Move(tempPath, destPath); err != nil {
		return preparedImage{}, fmt.Errorf("move resized image: %w", err)
	}

	relative := path.Join(postImagesDir, postImagesFolder, fileName)

	return preparedImage{
		URL:  g.siteURLFor(relative),
		Mime: pkgimages.MIMEFromExtension(ext),
	}, nil
}

func (g *Generator) siteURLFor(rel string) string {
	base := strings.TrimSuffix(g.Page.SiteURL, "/")
	rel = strings.TrimPrefix(rel, "/")

	if base == "" {
		return rel
	}

	return portal.SanitiseURL(base + "/" + rel)
}
