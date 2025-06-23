package handler

import (
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type Talks struct {
	content string
}

func MakeTalks() Talks {
	return Talks{
		content: "./storage/fixture/talks.json",
	}
}

func (h Talks) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.content)

	if err != nil {
		slog.Error("Error reading talks file: %v", err)

		return http.InternalError("could not read talks data")
	}

	if err := writeJSON(fixture, w); err != nil {
		return http.InternalError(err.Error())
	}

	return nil // A nil return indicates success.
}
