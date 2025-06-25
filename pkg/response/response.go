package response

import (
	"encoding/json"
	"fmt"
	baseHttp "net/http"
	"strings"
)

type Response struct {
	version string
	etag    string
	writer  baseHttp.ResponseWriter
	request *baseHttp.Request
}

func MakeFrom(version string, w baseHttp.ResponseWriter, r *baseHttp.Request) Response {
	v := strings.TrimSpace(version)

	return Response{
		version: v,
		writer:  w,
		request: r,
		etag:    fmt.Sprintf(`"%s"`, v),
	}
}

func (r *Response) Encode(payload any) error {
	return json.NewEncoder(r.writer).Encode(payload)
}

func (r *Response) HasCache() bool {
	request := r.request

	match := strings.TrimSpace(
		request.Header.Get("If-None-Match"),
	)

	return match == r.etag
}

func (r *Response) SetHeaders() {
	w := r.writer

	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", r.etag)
}

func (r *Response) RespondWithNotModified() {
	r.writer.WriteHeader(baseHttp.StatusNotModified)
}

func (r *Response) RespondOk() {
	r.writer.WriteHeader(baseHttp.StatusOK)
}
