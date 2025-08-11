package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"

	"log/slog"
	baseHttp "net/http"
)

type ProfileHandler struct {
	filePath string
}

func MakeProfileHandler(filePath string) ProfileHandler {
	return ProfileHandler{
		filePath: filePath,
	}
}

func (h ProfileHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	data, err := portal.ParseJsonFile[payload.ProfileResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading profile file", "error", err)

		return http.InternalError("could not read profile data")
	}

	resp := http.MakeResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for profile response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
