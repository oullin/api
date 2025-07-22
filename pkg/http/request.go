package http

import (
	"encoding/json"
	"fmt"
	"io"
	baseHttp "net/http"
)

const MaxRequestSize = 1 << 20 // 1MB limit

func ParseRequestBody[T any](r *baseHttp.Request) (T, error) {
	var err error
	var request T
	var data []byte

	limitedReader := io.LimitReader(r.Body, MaxRequestSize)
	if data, err = io.ReadAll(limitedReader); err != nil {
		return request, fmt.Errorf("failed to read the given request body: %w", err)
	}

	if len(data) == 0 {
		return request, nil
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return request, fmt.Errorf("failed to unmarshal the given request body: %w", err)
	}

	return request, nil
}
