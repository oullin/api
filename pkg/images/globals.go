package images

import "io"

const DefaultJPEGQuality = 85

type composedReadCloser struct {
	io.Reader
	io.Closer
}
