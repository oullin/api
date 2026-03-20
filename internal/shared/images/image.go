package images

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	stdimage "image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"github.com/andybalholm/brotli"
	"github.com/gen2brain/avif"
	"github.com/klauspost/compress/zstd"
)

type fetchRequest struct {
	URL    *url.URL
	Accept string
}

func (r fetchRequest) key() string {
	if r.URL == nil {
		return r.Accept
	}

	return r.Accept + " " + r.URL.String()
}

func Fetch(source string) (stdimage.Image, string, error) {
	parsed, err := url.Parse(source)
	if err != nil {
		return nil, "", fmt.Errorf("parse url: %w", err)
	}

	queue := []fetchRequest{{URL: parsed, Accept: supportedImageAcceptHeader}}
	seen := make(map[string]struct{})

	var (
		lastErr         error
		lastDecodeErr   error
		lastPayload     []byte
		lastContentType string
		lastEncoding    string
	)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.URL == nil {
			continue
		}

		key := current.key()
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		reader, contentType, encoding, err := openSource(current)
		if err != nil {
			lastErr = err
			continue
		}

		payload, err := readImagePayload(reader)
		if err != nil {
			lastErr = err
			continue
		}

		img, format, err := decodeImagePayload(payload)
		if err == nil {
			return img, format, nil
		}

		lastDecodeErr = err
		lastPayload = payload
		lastContentType = contentType
		lastEncoding = encoding

		queue = append(queue, githubAttachmentFallbacks(current, payload)...)
	}

	if lastDecodeErr != nil {
		return nil, "", newDecodeError(lastDecodeErr, lastPayload, lastContentType, lastEncoding)
	}

	if lastErr != nil {
		return nil, "", lastErr
	}

	return nil, "", errors.New("failed to fetch image")
}

func readImagePayload(reader io.ReadCloser) ([]byte, error) {
	defer reader.Close()

	limited := io.LimitReader(reader, maxRemoteImageBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read image payload: %w", err)
	}

	if len(data) == 0 {
		return nil, errors.New("empty image payload")
	}

	if len(data) > maxRemoteImageBytes {
		return nil, fmt.Errorf("image payload exceeds %d bytes", maxRemoteImageBytes)
	}

	return data, nil
}

func decodeImagePayload(data []byte) (stdimage.Image, string, error) {
	queue := [][]byte{data}
	seen := make(map[[32]byte]struct{})

	var lastErr error

	for len(queue) > 0 {
		candidate := queue[0]
		queue = queue[1:]

		hash := sha256.Sum256(candidate)
		if _, exists := seen[hash]; exists {
			continue
		}
		seen[hash] = struct{}{}

		img, format, err := stdimage.Decode(bytes.NewReader(candidate))
		if err == nil {
			return img, format, nil
		}

		lastErr = err

		if isAVIF(candidate) {
			avifImg, avifErr := decodeAVIF(candidate)
			if avifErr == nil {
				return avifImg, "avif", nil
			}

			lastErr = fmt.Errorf("decode avif payload: %w", avifErr)
		}

		trimmed := trimLeadingNoise(candidate)
		if len(trimmed) > 0 && len(trimmed) != len(candidate) {
			queue = append(queue, trimmed)
		}

		if start, ok := findEmbeddedImageStart(candidate); ok && start > 0 && start < len(candidate) {
			queue = append(queue, candidate[start:])
		}

		queue = append(queue, expandCompressedCandidate(candidate)...)
	}

	if lastErr == nil {
		lastErr = errors.New("image: unknown format")
	}

	return nil, "", lastErr
}

func trimLeadingNoise(data []byte) []byte {
	trimmed := dropUTF8BOM(data)
	trimmed = bytes.TrimLeftFunc(trimmed, unicode.IsSpace)

	return dropUTF8BOM(trimmed)
}

func dropUTF8BOM(data []byte) []byte {
	for len(data) >= len(utf8BOM) && bytes.Equal(data[:len(utf8BOM)], utf8BOM) {
		data = data[len(utf8BOM):]
	}

	return data
}

func expandCompressedCandidate(data []byte) [][]byte {
	var expansions [][]byte

	if decoded, err := tryBrotliDecode(data); err == nil {
		expansions = append(expansions, decoded)
	}

	if decoded, err := tryGzipDecode(data); err == nil {
		expansions = append(expansions, decoded)
	}

	if decoded, err := tryZlibDecode(data); err == nil {
		expansions = append(expansions, decoded)
	}

	if decoded, err := tryZstdDecode(data); err == nil {
		expansions = append(expansions, decoded)
	}

	return expansions
}

func isAVIF(data []byte) bool {
	if len(data) < 16 {
		return false
	}

	boxSize := binary.BigEndian.Uint32(data[:4])
	if boxSize == 0 || int(boxSize) > len(data) {
		boxSize = uint32(len(data))
	}

	if string(data[4:8]) != "ftyp" {
		return false
	}

	brands := [][]byte{data[8:12]}
	for offset := 16; offset+4 <= int(boxSize); offset += 4 {
		brands = append(brands, data[offset:offset+4])
	}

	for _, brand := range brands {
		switch string(brand) {
		case "avif", "avis", "avio":
			return true
		}
	}

	return false
}

func decodeAVIF(data []byte) (stdimage.Image, error) {
	avifInitOnce.Do(func() {
		avif.InitDecoder()
	})

	img, err := avif.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func githubAttachmentFallbacks(current fetchRequest, payload []byte) []fetchRequest {
	u := current.URL
	if u == nil || len(payload) == 0 {
		return nil
	}

	if !isAVIF(payload) {
		return nil
	}

	id, ok := githubAttachmentID(u)
	if !ok {
		return nil
	}

	basePaths := []*url.URL{
		{Scheme: u.Scheme, Host: u.Host, Path: path.Join("/user-attachments/assets", id)},
		{Scheme: "https", Host: "github.com", Path: path.Join("/user-attachments/assets", id)},
	}

	variants := make([]fetchRequest, 0, len(basePaths)*5)

	for _, base := range basePaths {
		if base.Host == "" {
			continue
		}

		for _, option := range []struct {
			format string
			name   string
		}{
			{format: "png", name: "large"},
			{format: "png", name: "medium"},
			{format: "jpg", name: "large"},
			{format: "jpg", name: "medium"},
		} {
			clone := *base
			query := clone.Query()
			query.Set("format", option.format)
			query.Set("name", option.name)
			clone.RawQuery = query.Encode()
			accept := current.Accept
			switch option.format {
			case "png":
				accept = fallbackPNGAcceptHeader
			case "jpg":
				accept = fallbackJPEGAcceptHeader
			}

			variants = append(variants, fetchRequest{URL: &clone, Accept: accept})
		}

		clone := *base
		query := clone.Query()
		query.Set("raw", "1")
		clone.RawQuery = query.Encode()
		variants = append(variants, fetchRequest{URL: &clone, Accept: current.Accept})
	}

	return variants
}

func githubAttachmentID(u *url.URL) (string, bool) {
	if u == nil {
		return "", false
	}

	trimmedPath := strings.Trim(u.Path, "/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) >= 3 && parts[0] == "user-attachments" && parts[1] == "assets" {
		id := parts[2]
		if id != "" {
			return id, true
		}
	}

	if strings.Contains(strings.ToLower(u.Host), "github-production-user-asset") {
		base := path.Base(u.Path)
		if base == "" {
			return "", false
		}

		if dot := strings.LastIndex(base, "."); dot >= 0 {
			base = base[:dot]
		}

		if dash := strings.Index(base, "-"); dash >= 0 && dash+1 < len(base) {
			base = base[dash+1:]
		}

		base = strings.TrimSpace(base)
		if base != "" {
			return base, true
		}
	}

	return "", false
}

func tryBrotliDecode(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))

	decoded, err := readLimited(reader, maxRemoteImageBytes)
	if err != nil {
		return nil, err
	}

	if len(decoded) == 0 {
		return nil, errors.New("brotli decoded empty")
	}

	return decoded, nil
}

func tryGzipDecode(data []byte) ([]byte, error) {
	if len(data) < 2 || data[0] != 0x1F || data[1] != 0x8B {
		return nil, errors.New("not gzip")
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return readLimited(reader, maxRemoteImageBytes)
}

func tryZlibDecode(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, errors.New("not zlib")
	}

	cmf := data[0]
	flg := data[1]

	if cmf&0x0F != 8 { // compression method deflate
		return nil, errors.New("not zlib deflate")
	}

	if (uint16(cmf)<<8|uint16(flg))%31 != 0 {
		return nil, errors.New("invalid zlib header")
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return readLimited(reader, maxRemoteImageBytes)
}

func tryZstdDecode(data []byte) ([]byte, error) {
	if len(data) < 4 || data[0] != 0x28 || data[1] != 0xB5 || data[2] != 0x2F || data[3] != 0xFD {
		return nil, errors.New("not zstd")
	}

	decoder, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	return readLimited(decoder, maxRemoteImageBytes)
}

func readLimited(reader io.Reader, limit int) ([]byte, error) {
	limited := io.LimitReader(reader, int64(limit)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if len(data) > limit {
		return nil, fmt.Errorf("decompressed payload exceeds %d bytes", limit)
	}

	if len(data) == 0 {
		return nil, errors.New("decompressed payload empty")
	}

	return data, nil
}

func findEmbeddedImageStart(data []byte) (int, bool) {
	if idx := bytes.Index(data, []byte{0xFF, 0xD8, 0xFF}); idx >= 0 {
		return idx, true
	}

	if idx := bytes.Index(data, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}); idx >= 0 {
		return idx, true
	}

	if idx := bytes.Index(data, []byte("GIF87a")); idx >= 0 {
		return idx, true
	}

	if idx := bytes.Index(data, []byte("GIF89a")); idx >= 0 {
		return idx, true
	}

	for idx := bytes.Index(data, []byte("RIFF")); idx >= 0; {
		if len(data)-idx >= 12 && bytes.Equal(data[idx+8:idx+12], []byte("WEBP")) {
			return idx, true
		}

		next := bytes.Index(data[idx+4:], []byte("RIFF"))
		if next < 0 {
			break
		}

		idx += 4 + next
	}

	return 0, false
}

func Resize(src stdimage.Image, width, height int) stdimage.Image {
	dst := stdimage.NewRGBA(stdimage.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return dst
}

func DetermineExtension(source, format string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(source)))
	format = strings.ToLower(strings.TrimSpace(format))

	switch ext {
	case ".jpeg":
		return ".jpg"
	case ".jpg", ".png", ".webp":
		return ext
	}

	switch format {
	case "jpeg", "jpg":
		return ".jpg"
	case "png":
		return ".png"
	case "webp":
		return ".webp"
	}

	return ".jpg"
}

func BuildFileName(slug, ext, fallback string) string {
	trimmed := strings.TrimSpace(slug)
	cleaned := strings.Trim(trimmed, "/")
	if cleaned == "" {
		cleaned = fallback
	}

	cleaned = strings.ReplaceAll(cleaned, " ", "-")

	return cleaned + ext
}

func Save(path string, img stdimage.Image, ext string, quality int) error {
	fh, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fh.Close()

	switch ext {
	case ".png":
		encoder := &png.Encoder{CompressionLevel: png.DefaultCompression}
		return encoder.Encode(fh, img)
	case ".webp":
		return encodeWebp(fh, img, quality)
	default:
		options := &jpeg.Options{Quality: quality}
		return jpeg.Encode(fh, img, options)
	}
}

func Move(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	return os.Remove(src)
}

func MIMEFromExtension(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func NormalizeRelativeURL(rel string) string {
	rel = strings.ReplaceAll(rel, "\\", "/")

	cleaned := path.Clean(rel)

	if cleaned == "." || cleaned == "/" {
		return ""
	}

	parts := strings.Split(cleaned, "/")

	var b strings.Builder

	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}

		if b.Len() > 0 {
			b.WriteByte('/')
		}

		b.WriteString(part)
	}

	return b.String()
}

func openSource(req fetchRequest) (io.ReadCloser, string, string, error) {
	parsed := req.URL
	switch parsed.Scheme {
	case "http", "https":
		client := &http.Client{Timeout: 10 * time.Second}

		httpReq, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
		if err != nil {
			return nil, "", "", fmt.Errorf("create request: %w", err)
		}

		accept := strings.TrimSpace(req.Accept)
		if accept == "" {
			accept = supportedImageAcceptHeader
		}

		httpReq.Header.Set("Accept", accept)
		httpReq.Header.Set("User-Agent", defaultRemoteImageUA)

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, "", "", fmt.Errorf("download image: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			defer resp.Body.Close()
			return nil, "", "", fmt.Errorf("download image: unexpected status %s", resp.Status)
		}

		reader, encoding, err := wrapHTTPBody(resp)
		if err != nil {
			return nil, "", "", err
		}

		return reader, resp.Header.Get("Content-Type"), encoding, nil
	case "file":
		reader, err := openLocal(parsed)
		return reader, "", "", err
	case "":
		reader, err := os.Open(parsed.Path)
		return reader, "", "", err
	default:
		return nil, "", "", fmt.Errorf("unsupported image scheme: %s", parsed.Scheme)
	}
}

func wrapHTTPBody(resp *http.Response) (io.ReadCloser, string, error) {
	encoding := strings.TrimSpace(strings.ToLower(resp.Header.Get("Content-Encoding")))
	if idx := strings.IndexRune(encoding, ','); idx >= 0 {
		encoding = encoding[:idx]
	}
	switch encoding {
	case "", "identity":
		return resp.Body, encoding, nil
	case "br":
		return composedReadCloser{Reader: brotli.NewReader(resp.Body), Closer: resp.Body}, encoding, nil
	case "gzip":
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, encoding, fmt.Errorf("prepare gzip decoder: %w", err)
		}

		return composedReadCloser{Reader: reader, Closer: multiCloser{reader, resp.Body}}, encoding, nil
	case "zstd", "zstandard":
		decoder, err := zstd.NewReader(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, encoding, fmt.Errorf("prepare zstd decoder: %w", err)
		}

		return composedReadCloser{Reader: decoder, Closer: multiCloser{noErrorCloseFunc(decoder.Close), resp.Body}}, encoding, nil
	default:
		return resp.Body, encoding, nil
	}
}

func openLocal(parsed *url.URL) (io.ReadCloser, error) {
	pathValue := parsed.Path
	if pathValue == "" {
		pathValue = parsed.Opaque
	}

	if parsed.Host != "" {
		pathValue = "//" + parsed.Host + pathValue
	}

	unescaped, err := url.PathUnescape(pathValue)
	if err != nil {
		return nil, fmt.Errorf("decode file path: %w", err)
	}

	return os.Open(unescaped)
}
