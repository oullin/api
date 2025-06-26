package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
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
	data, err := pkg.ParseJsonFile[payload.ProfileResponse](h.fixture)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return http.InternalError("could not read profile data")
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
