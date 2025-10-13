//go:build !cgo

package images

import (
	"errors"
	stdimage "image"
	"io"
)

func encodeWebp(w io.Writer, img stdimage.Image, quality int) error {
	return errors.New("webp encoding requires cgo support")
}

func webpEncodeSupported() bool {
	return false
}
