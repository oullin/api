package pkg

import (
	"io"
	"log/slog"
)

func CloseWithLog(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Error("failed to close resource", "err", err)
	}
}
