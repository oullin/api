package seo

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/cli"
	pkgimages "github.com/oullin/pkg/images"
	"github.com/oullin/pkg/portal"
)

type preparedImage struct {
	URL  string
	Mime string
}

const (
	seoImageWidth  = 1200
	seoImageHeight = 630
)

func (g *Generator) preparePostImage(post payload.PostResponse) (preparedImage, error) {
	source := strings.TrimSpace(post.CoverImageURL)
	if source == "" {
		cli.Grayln(fmt.Sprintf("Skipping image preparation for %s: no cover image", post.Slug))
		return preparedImage{}, errors.New("post has no cover image url")
	}

	cli.Grayln(fmt.Sprintf("Preparing post image for %s from %s", post.Slug, source))

	spaImagesDir, err := g.spaImagesDir()
	if err != nil {
		cli.Errorln(fmt.Sprintf("Could not determine SPA images directory: %v", err))
		return preparedImage{}, err
	}

	img, format, err := pkgimages.Fetch(source)
	if err != nil {
		var decodeErr *pkgimages.DecodeError
		if errors.As(err, &decodeErr) {
			cli.Errorln("Failed to decode remote image. Diagnostics:")
			for _, line := range decodeErr.Diagnostics() {
				cli.Grayln("  " + line)
			}
		}

		return preparedImage{}, err
	}

	bounds := img.Bounds()
	cli.Grayln(fmt.Sprintf("Fetched image %s (%s) with bounds %dx%d", source, format, bounds.Dx(), bounds.Dy()))

	resized := pkgimages.Resize(img, seoImageWidth, seoImageHeight)
	cli.Grayln(fmt.Sprintf("Resized image to %dx%d", seoImageWidth, seoImageHeight))

	ext := pkgimages.DetermineExtension(source, format)
	fileName := pkgimages.BuildFileName(post.Slug, ext, "post-image")

	if err := os.MkdirAll(spaImagesDir, 0o755); err != nil {
		cli.Errorln(fmt.Sprintf("Unable to create SPA images directory %s: %v", spaImagesDir, err))
		return preparedImage{}, fmt.Errorf("create destination dir: %w", err)
	}

	destPath := filepath.Join(spaImagesDir, fileName)
	if err := pkgimages.Save(destPath, resized, ext, pkgimages.DefaultJPEGQuality); err != nil {
		cli.Errorln(fmt.Sprintf("Unable to save resized image to %s: %v", destPath, err))
		return preparedImage{}, fmt.Errorf("write resized image: %w", err)
	}

	relativeDir, err := filepath.Rel(g.Page.OutputDir, spaImagesDir)
	if err != nil {
		cli.Errorln(fmt.Sprintf("Unable to determine relative path for %s: %v", spaImagesDir, err))
		return preparedImage{}, fmt.Errorf("determine relative image path: %w", err)
	}

	relativeDir = filepath.ToSlash(relativeDir)
	relativeDir = strings.Trim(relativeDir, "/")

	relative := path.Join(relativeDir, fileName)
	relative = pkgimages.NormalizeRelativeURL(relative)

	cli.Grayln(fmt.Sprintf("Post image relative path: %s", relative))

	return preparedImage{
		URL:  g.siteURLFor(relative),
		Mime: pkgimages.MIMEFromExtension(ext),
	}, nil
}

func (g *Generator) spaImagesDir() (string, error) {
	if g.Env == nil {
		return "", errors.New("generator environment not configured")
	}

	if g.Env.Seo.SpaImagesDir == "" {
		return "", errors.New("spa images directory is not configured")
	}

	return g.Env.Seo.SpaImagesDir, nil
}

func (g *Generator) siteURLFor(rel string) string {
	base := strings.TrimSuffix(g.Page.SiteURL, "/")
	rel = strings.TrimPrefix(rel, "/")

	if base == "" {
		return rel
	}

	return portal.SanitiseURL(base + "/" + rel)
}
