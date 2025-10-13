//go:build !cgo

package images

import (
	"errors"
	stdimage "image"
	"io"
)

func encodeWebp(_ io.Writer, _ stdimage.Image, _ int) error {
	return errors.New("webp encoding requires cgo support")
}

func webpEncodeSupported() bool {
	return false
}
