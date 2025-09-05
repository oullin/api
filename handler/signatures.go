package handler

import (
	"encoding/json"
	"fmt"
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
		return http.LogBadRequestError("could not read signatures request body", err)
	}

	var req payload.SignatureRequest
	if err = json.Unmarshal(bodyBytes, &req); err != nil {
		return http.LogBadRequestError("could not parse the given data.", err)
	}

	if _, err = s.Validator.Rejects(req); err != nil {
		return http.UnprocessableEntity("The given fields are invalid", s.Validator.GetErrors())
	}

	serverTime := time.Now()
	receivedAt := time.Unix(req.Timestamp, 0)

	if err = s.isRequestWithinTimeframe(serverTime, receivedAt); err != nil {
		return http.LogBadRequestError(err.Error(), err)
	}

	var keySignature *database.APIKeySignatures
	if keySignature, err = s.CreateSignature(req.Username, serverTime); err != nil {
		return http.LogInternalError(err.Error(), err)
	}

	response := payload.SignatureResponse{
		Signature: auth.SignatureToString(keySignature.Signature),
		Tries:     keySignature.Tries,
		Cadence: payload.SignatureCadenceResponse{
			ReceivedAt: receivedAt.Format(portal.DatesLayout),
			CreatedAt:  keySignature.CreatedAt.Format(portal.DatesLayout),
			ExpiresAt:  keySignature.ExpiresAt.Format(portal.DatesLayout),
		},
	}

	resp := http.MakeResponseFrom("0.0.1", w, r)

	if err = resp.RespondOk(response); err != nil {
		slog.Error("Error marshaling JSON for signatures response", "error", err)
		return nil
	}

	return nil // A nil return indicates success.
}

func (s *SignaturesHandler) isRequestWithinTimeframe(serverTime, receivedAt time.Time) error {
	//skew := 5 * time.Second

	//earliestValidTime := serverTime.Add(-skew)
	//if receivedAt.Before(earliestValidTime) {
	//	return fmt.Errorf("the request timestamp [%s] is too old", receivedAt.Format(portal.DatesLayout))
	//}
	//
	//latestValidTime := serverTime.Add(skew)
	//if receivedAt.After(latestValidTime) {
	//	return fmt.Errorf("the request timestamp [%s] is from the future", receivedAt.Format(portal.DatesLayout))
	//}

	return nil
}

func (s *SignaturesHandler) CreateSignature(username string, serverTime time.Time) (*database.APIKeySignatures, error) {
	var err error
	var token *database.APIKey
	var keySignature *database.APIKeySignatures

	if token = s.ApiKeys.FindBy(username); token == nil {
		return nil, fmt.Errorf("the given username [%s] was not found", username)
	}

	var seed []byte
	if seed, err = auth.GenerateAESKey(); err != nil {
		return nil, fmt.Errorf("unable to generate the signature seed. Please try again")
	}

	expiresAt := serverTime.Add(time.Second * 30)
	hash := auth.CreateSignature(seed, token.SecretKey)

	if keySignature, err = s.ApiKeys.CreateSignatureFor(token, hash, expiresAt); err != nil {
		return nil, fmt.Errorf("unable to create the signature item. Please try again")
	}

	return keySignature, nil
}
