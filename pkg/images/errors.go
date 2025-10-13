package images

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
)

type DecodeError struct {
	Err             error
	ContentType     string
	ContentEncoding string
	SniffedType     string
	Size            int
	Hash            string
	PrefixHex       string
}

func newDecodeError(err error, payload []byte, contentType, encoding string) *DecodeError {
	sample := payload
	if len(sample) > 512 {
		sample = sample[:512]
	}

	sniffed := http.DetectContentType(sample)

	details := &DecodeError{
		Err:             err,
		ContentType:     strings.TrimSpace(contentType),
		ContentEncoding: strings.TrimSpace(encoding),
		SniffedType:     sniffed,
		Size:            len(payload),
	}

	if len(payload) > 0 {
		sum := sha256.Sum256(payload)
		details.Hash = hex.EncodeToString(sum[:])

		prefixLen := len(payload)
		if prefixLen > 16 {
			prefixLen = 16
		}

		details.PrefixHex = hex.EncodeToString(payload[:prefixLen])
	}

	return details
}

func (e *DecodeError) Error() string {
	if e == nil {
		return "decode image: <nil>"
	}

	var parts []string

	if ct := strings.TrimSpace(e.ContentType); ct != "" {
		parts = append(parts, fmt.Sprintf("content-type %q", ct))
	}

	if enc := strings.TrimSpace(e.ContentEncoding); enc != "" {
		parts = append(parts, fmt.Sprintf("content-encoding %q", enc))
	}

	if sniff := strings.TrimSpace(e.SniffedType); sniff != "" {
		parts = append(parts, fmt.Sprintf("sniffed %q", sniff))
	}

	if e.Size > 0 {
		parts = append(parts, fmt.Sprintf("size %d bytes", e.Size))
	}

	if e.Hash != "" {
		parts = append(parts, fmt.Sprintf("sha256 %s", e.Hash))
	}

	if e.PrefixHex != "" {
		parts = append(parts, fmt.Sprintf("prefix %s", e.PrefixHex))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("decode image: %v", e.Err)
	}

	return fmt.Sprintf("decode image (%s): %v", strings.Join(parts, ", "), e.Err)
}

func (e *DecodeError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *DecodeError) Diagnostics() []string {
	if e == nil {
		return nil
	}

	var lines []string

	if ct := strings.TrimSpace(e.ContentType); ct != "" {
		lines = append(lines, fmt.Sprintf("content-type: %s", ct))
	} else {
		lines = append(lines, "content-type: (missing)")
	}

	if enc := strings.TrimSpace(e.ContentEncoding); enc != "" {
		lines = append(lines, fmt.Sprintf("content-encoding: %s", enc))
	} else {
		lines = append(lines, "content-encoding: (missing)")
	}

	if sniff := strings.TrimSpace(e.SniffedType); sniff != "" {
		lines = append(lines, fmt.Sprintf("sniffed-type: %s", sniff))
	}

	lines = append(lines, fmt.Sprintf("payload-bytes: %d", e.Size))

	if e.Hash != "" {
		lines = append(lines, fmt.Sprintf("payload-sha256: %s", e.Hash))
	}

	if e.PrefixHex != "" {
		lines = append(lines, fmt.Sprintf("payload-prefix-hex: %s", e.PrefixHex))
	}

	return lines
}
