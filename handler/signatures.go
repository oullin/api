package handler

import (
	"encoding/json"
	"log/slog"
	baseHttp "net/http"

	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"
)

type SignatureRequest struct {
	//Method      string `json:"method"`
	//URL         string `json:"url"`
	//Body        string `json:"body"`
	Nonce       string `json:"nonce" validate:"required,lowercase,len=32"`
	APIKey      string `json:"apiKey" validate:"required,lowercase,len=32"`
	APIUsername string `json:"apiUsername" validate:"required,lowercase,min=5"`
	Timestamp   string `json:"timestamp" validate:"required"`
}

type SignatureResponse struct {
	Signature string `json:"signature"`
}

type SignaturesHandler struct {
	validator *portal.Validator
}

func MakeSignaturesHandler(validator *portal.Validator) SignaturesHandler {
	return SignaturesHandler{
		validator: validator,
	}
}

func (s *SignaturesHandler) Generate(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	defer portal.CloseWithLog(r.Body)

	var req SignatureRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Error reading signatures request", "error", err)

		return http.InternalError("could not read signatures data")
	}

	signature := SignatureResponse{Signature: "TEST"}
	resp := http.MakeResponseFrom("0.0.1", w, r)

	if err := resp.RespondOk(signature); err != nil {
		slog.Error("Error marshaling JSON for signatures response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
