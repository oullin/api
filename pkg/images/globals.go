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

type multiCloser []io.Closer

func (m multiCloser) Close() error {
	var firstErr error

	for _, closer := range m {
		if closer == nil {
			continue
		}

		if err := closer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

type noErrorCloseFunc func()

func (f noErrorCloseFunc) Close() error {
	if f == nil {
		return nil
	}

	f()

	return nil
}
