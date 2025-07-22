package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	baseHttp "net/http"
)

func ParseRequestBody[T any](r *baseHttp.Request) (T, func(), error) {
	var err error
	var request T
	var data []byte

	closer := func() {
		defer func(Body io.ReadCloser) {
			if issue := Body.Close(); issue != nil {
				slog.Error("ParseRequestBody: " + issue.Error())
			}
		}(r.Body)
	}

	if data, err = io.ReadAll(r.Body); err != nil {
		return request, closer, fmt.Errorf("failed to read the given request body: %w", err)
	}

	if len(data) == 0 {
		return request, closer, nil
	}

	if err = json.Unmarshal(data, &request); err != nil {
		return request, closer, fmt.Errorf("failed to unmarshal the given request body: %w", err)
	}

	return request, closer, nil
}
