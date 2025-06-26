package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

type ProfileHandler struct {
	filePah string
}

func MakeProfileHandler(filePah string) ProfileHandler {
	return ProfileHandler{
		filePah: filePah,
	}
}

func (h ProfileHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	data, err := pkg.ParseJsonFile[payload.ProfileResponse](h.filePah)

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
