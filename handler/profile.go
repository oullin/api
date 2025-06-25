package handler

import (
	"encoding/json"
	"fmt"
	"github.com/oullin/handler/response"
	"github.com/oullin/pkg/http"
	"log"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type ProfileHandler struct {
	content string
}

func MakeProfileHandler(fixture string) ProfileHandler {
	return ProfileHandler{
		content: fixture,
	}
}

func (h ProfileHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.content)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return http.InternalError("could not read profile data")
	}

	var data response.ProfileResponse
	if err := json.Unmarshal(fixture, &data); err != nil {
		return http.InternalError(err.Error())
	}

	version := data.Version
	etag := fmt.Sprintf(`"%s"`, version)

	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			// If the ETags match, the client's version is fresh.
			// Send back a 304 Not Modified status and an empty body.
			w.WriteHeader(baseHttp.StatusNotModified)

			return nil
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", etag)
	w.WriteHeader(baseHttp.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// This error could happen if the struct has unmarshallable types (e.g., channels).
		log.Printf("Error marshalling JSON for response: %v", err)
		// The header might already be sent, so we can't send a new http.Error.
		// We just log the error.
		return nil
	}

	// Marshal and send the JSON data
	//json.NewEncoder(w).Encode(responseData)

	//if err := writeJSON(fixture, w); err != nil {
	//	return http.InternalError(err.Error())
	//}

	return nil // A nil return indicates success.
}
