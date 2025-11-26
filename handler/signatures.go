package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/repoentity"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
)

type SignaturesHandler struct {
	Validator *portal.Validator
	ApiKeys   *repository.ApiKeys
}

func NewSignaturesHandler(validator *portal.Validator, ApiKeys *repository.ApiKeys) SignaturesHandler {
	return SignaturesHandler{
		Validator: validator,
		ApiKeys:   ApiKeys,
	}
}

func (s *SignaturesHandler) Generate(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	defer portal.CloseWithLog(r.Body)

	var (
		err error
		req payload.SignatureRequest
	)

	r.Body = http.MaxBytesReader(w, r.Body, endpoint.MaxRequestSize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err = dec.Decode(&req); err != nil {
		return endpoint.LogBadRequestError("could not parse the given data.", err)
	}

	if _, err = s.Validator.Rejects(req); err != nil {
		return endpoint.UnprocessableEntity("The given fields are invalid", s.Validator.GetErrors())
	}

	serverTime := time.Now()
	receivedAt := time.Unix(req.Timestamp, 0)
	req.Origin = portal.IntendedOriginFromHeader(r.Header)

	var keySignature *database.APIKeySignatures
	if keySignature, err = s.CreateSignature(req, serverTime); err != nil {
		return endpoint.LogInternalError(err.Error(), err)
	}

	response := payload.SignatureResponse{
		Signature: auth.SignatureToString(keySignature.Signature),
		MaxTries:  keySignature.MaxTries,
		Cadence: payload.SignatureCadenceResponse{
			ReceivedAt: receivedAt.Format(portal.DatesLayout),
			CreatedAt:  keySignature.CreatedAt.Format(portal.DatesLayout),
			ExpiresAt:  keySignature.ExpiresAt.Format(portal.DatesLayout),
		},
	}

	resp := endpoint.NewResponseFrom("0.0.1", w, r)

	if err = resp.RespondOk(response); err != nil {
		slog.Error("Error marshaling JSON for signatures response", "error", err)
		return endpoint.LogInternalError("could not encode signatures response", err)
	}

	return nil // A nil return indicates success.
}

func (s *SignaturesHandler) CreateSignature(request payload.SignatureRequest, serverTime time.Time) (*database.APIKeySignatures, error) {
	var err error
	var token *database.APIKey
	var keySignature *database.APIKeySignatures

	if token = s.ApiKeys.FindBy(request.Username); token == nil {
		return nil, fmt.Errorf("the given username [%s] was not found", request.Username)
	}

	var seed []byte
	if seed, err = auth.GenerateAESKey(); err != nil {
		return nil, fmt.Errorf("unable to generate the signature seed. Please try again")
	}

	// Signature TTL: 5 minutes (300 seconds)
	//
	// Security model:
	// - Each signature can only be used once (enforced by nonce validation)
	// - Nonces are cached for 25 minutes to prevent replay attacks
	// - Max 10 validation attempts per signature prevents brute force
	// - Rate limiter allows 50 failures per 5-minute window per IP+account
	//
	// The 5-minute window is conservative and industry-standard:
	// - AWS Signature V4: 15 minutes
	// - Azure SAS: configurable, typically 15-60 minutes
	// - Google Cloud: 15 minutes
	//
	// This allows adequate time for:
	// - Network latency (typically < 1 second)
	// - Client processing/retries (< 30 seconds typical)
	// - Clock skew between client/server (we allow ±10 minutes)
	expiresAt := serverTime.Add(time.Minute * 5)
	hash := auth.CreateSignature(seed, token.SecretKey)

	entity := repoentity.APIKeyCreateSignatureFor{
		Key:       token,
		ExpiresAt: expiresAt,
		Seed:      hash,
		Origin:    portal.NormalizeOriginWithPath(request.Origin),
	}

	if keySignature, err = s.ApiKeys.CreateSignatureFor(entity); err != nil {
		return nil, fmt.Errorf("unable to create the signature item. Please try again")
	}

	return keySignature, nil
}
