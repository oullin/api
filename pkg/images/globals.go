package images

import "io"

const (
	DefaultJPEGQuality         = 85
	supportedImageAcceptHeader = "image/webp,image/png,image/jpeg,image/gif;q=0.9,*/*;q=0.1"
	defaultRemoteImageUA       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.2 Safari/605.1.15"
)

type composedReadCloser struct {
	io.Reader
	io.Closer
}
