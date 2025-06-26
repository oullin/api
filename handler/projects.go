package handler

import (
    "github.com/oullin/handler/payload"
    "github.com/oullin/pkg"
    "github.com/oullin/pkg/http"
    "log/slog"
    baseHttp "net/http"
)

type ProjectsHandler struct {
    filePah string
}

func MakeProjectsHandler(filePah string) ProjectsHandler {
    return ProjectsHandler{
        filePah: filePah,
    }
}

func (h ProjectsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
    data, err := pkg.ParseJsonFile[payload.ProjectsResponse](h.filePah)

    if err != nil {
        slog.Error("Error reading projects file", "error", err)

        return http.InternalError("could not read projects data")
    }

    resp := http.MakeResponseFrom(data.Version, w, r)

    if resp.HasCache() {
        resp.RespondWithNotModified()

        return nil
    }

    if err := resp.RespondOk(data); err != nil {
        slog.Error("Error marshaling JSON for projects response", "error", err)

        return nil
    }

    return nil // A nil return indicates success.
}
