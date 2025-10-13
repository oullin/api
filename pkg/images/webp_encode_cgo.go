//go:build cgo

package images

import (
	stdimage "image"
	"io"

	"github.com/chai2010/webp"
)

func encodeWebp(w io.Writer, img stdimage.Image, quality int) error {
	options := &webp.Options{Lossless: false, Quality: float32(quality)}
	return webp.Encode(w, img, options)
}

func webpEncodeSupported() bool {
	return true
}
