package handler

import (
	"encoding/json"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type ProfileHandler struct {
	fixture string
}

func MakeProfileHandler(file string) ProfileHandler {
	return ProfileHandler{
		fixture: file,
	}
}

func (h ProfileHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.fixture)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return http.InternalError("could not read profile data")
	}

	var data payload.ProfileResponse
	if err := json.Unmarshal(fixture, &data); err != nil {
		return http.InternalError(err.Error())
	}

	resp := http.MakeResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
