package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	baseHttp "net/http"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"
)

type SignatureRequest struct {
	Method    string `json:"method" validate:"required,eq=POST"`
	URL       string `json:"url" validate:"required,uri"`
	Nonce     string `json:"nonce" validate:"required,lowercase,len=32"`
	PublicKey string `json:"public_key" validate:"required,lowercase,min=64,max=67"`
	Username  string `json:"username" validate:"required,lowercase,min=5"`
	Timestamp string `json:"timestamp" validate:"required,number,len=10"`
}

type SignatureResponse struct {
	Signature string `json:"signature"`
}

type SignaturesHandler struct {
	Validator *portal.Validator
	ApiKeys   *repository.ApiKeys
}

func MakeSignaturesHandler(validator *portal.Validator, ApiKeys *repository.ApiKeys) SignaturesHandler {
	return SignaturesHandler{
		Validator: validator,
		ApiKeys:   ApiKeys,
	}
}

func (s *SignaturesHandler) Generate(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	defer portal.CloseWithLog(r.Body)

	var req SignatureRequest

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading signatures request body", "error", err)

		return http.BadRequestError("could not read signatures request body")
	}

	if err = json.Unmarshal(bodyBytes, &req); err != nil {
		slog.Error("Error parsing signatures request", "error", err)

		return http.BadRequestError("could not parse the given data")
	}

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	req.Method = strings.ToUpper(r.Method)
	req.URL = portal.GenerateURL(r)

	//fmt.Println("-----> ", req)

	if _, err := s.Validator.Rejects(req); err != nil {
		return http.UnprocessableEntity("The given fields are invalid", s.Validator.GetErrors())
	}

	var token *database.APIKey
	if token = s.ApiKeys.FindBy(req.Username); token == nil {
		return http.NotFound("The given username was not found")
	}

	signature := SignatureResponse{Signature: "TEST"}
	resp := http.MakeResponseFrom("0.0.1", w, r)

	if err = resp.RespondOk(signature); err != nil {
		slog.Error("Error marshaling JSON for signatures response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
