package handler

import (
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type ProfileHandler struct {
	Fixture string
}

func (h ProfileHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.Fixture)

	if err != nil {
		slog.Error("Error reading profile file: %v", err)

		return &http.ApiError{
			Message: "Internal Server Error: could not read profile data",
			Status:  baseHttp.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(baseHttp.StatusOK)

	_, err = w.Write(fixture)

	if err != nil {
		slog.Error("Error writing response: %v", err)
	}

	return nil // A nil return indicates success.
}
