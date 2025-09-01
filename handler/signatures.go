package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	baseHttp "net/http"
	"net/url"
	"strings"

	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"
)

type SignatureRequest struct {
	Method      string `json:"method"`
	URL         string `json:"url"`
	Body        string `json:"body"`
	Nonce       string `json:"nonce" validate:"required,lowercase,len=32"`
	APIKey      string `json:"apiKey" validate:"required,lowercase,min=64,max=67"`
	APIUsername string `json:"apiUsername" validate:"required,lowercase,min=5"`
	Timestamp   string `json:"timestamp" validate:"required,lowercase,len=10"`
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

		return http.BadRequestError("could not read signatures data")
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading signatures request body", "error", err)

		return http.BadRequestError("could not read signatures request body")
	}

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		slog.Error("Error reading parsing the signature request URL", "error", err)

		return http.BadRequestError("could not read signatures URL")
	}

	req.Method = strings.ToUpper(req.Method)
	req.URL = parsedURL.String()
	req.Body = string(reqBody)

	if _, err := s.validator.Rejects(req); err != nil {
		return http.UnprocessableEntity("The given fields are invalid", s.validator.GetErrors())
	}

	signature := SignatureResponse{Signature: "TEST"}
	resp := http.MakeResponseFrom("0.0.1", w, r)

	if err := resp.RespondOk(signature); err != nil {
		slog.Error("Error marshaling JSON for signatures response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
