package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

type SocialHandler struct {
	filePah string
}

func MakeSocialHandler(filePah string) SocialHandler {
	return SocialHandler{
		filePah: filePah,
	}
}

func (h SocialHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	data, err := pkg.ParseJsonFile[payload.SocialResponse](h.filePah)

	if err != nil {
		slog.Error("Error reading social file", "error", err)

		return http.InternalError("could not read social data")
	}

	resp := http.MakeResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for social response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
