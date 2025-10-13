//go:build !cgo

package images

import (
	"errors"
	stdimage "image"
)

const avifSupported = false

func decodeAVIF(data []byte) (stdimage.Image, error) {
	return nil, errors.New("avif decoding requires cgo support")
}

func decodeAVIFConfig(data []byte) (stdimage.Config, error) {
	return stdimage.Config{}, errors.New("avif decoding requires cgo support")
}
