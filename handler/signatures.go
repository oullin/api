package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	baseHttp "net/http"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"
)

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

	var err error
	var bodyBytes []byte

	bodyBytes, err = io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Error reading signatures request body", "error", err)

		return http.BadRequestError("could not read signatures request body")
	}

	var req payload.SignatureRequest
	if err = json.Unmarshal(bodyBytes, &req); err != nil {
		slog.Error("Error parsing signatures request", "error", err)

		return http.BadRequestError("could not parse the given data.")
	}

	if _, err = s.Validator.Rejects(req); err != nil {
		return http.UnprocessableEntity("The given fields are invalid", s.Validator.GetErrors())
	}

	var token *database.APIKey
	if token = s.ApiKeys.FindBy(req.Username); token == nil {
		return http.NotFound("The given username was not found")
	}

	var seed []byte
	if seed, err = auth.GenerateAESKey(); err != nil {
		slog.Error("Error generating signatures seeds", "error", err)

		return http.InternalError("We were unable to generate the signature seed. Please try again!")
	}

	layout := "2006-01-02 15:04:05"
	receivedAt := time.Unix(req.Timestamp, 0)

	createdAt := time.Now()
	expiresAt := createdAt.Add(time.Second * 30)

	if receivedAt.Before(createdAt) {
		slog.Error("Invalid timestamp while creating signatures", "error", err)

		return http.BadRequestError("The given timestamp is before the current time")
	}

	response := payload.SignatureResponse{
		Signature: auth.CreateSignatureFrom(
			string(seed),
			string(token.SecretKey),
		),
		Cadence: payload.SignatureCadenceResponse{
			ReceivedAt: receivedAt.Format(layout),
			CreatedAt:  createdAt.Format(layout),
			ExpiresAt:  expiresAt.Format(layout),
		},
	}

	resp := http.MakeResponseFrom("0.0.1", w, r)
	if err = resp.RespondOk(response); err != nil {
		slog.Error("Error marshaling JSON for signatures response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
