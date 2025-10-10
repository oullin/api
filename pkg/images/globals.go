package images

import "io"

const (
	DefaultJPEGQuality         = 85
	supportedImageAcceptHeader = "image/webp,image/png,image/jpeg,image/gif;q=0.9,*/*;q=0.1"
)

type composedReadCloser struct {
	io.Reader
	io.Closer
}
