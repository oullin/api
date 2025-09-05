package payload

type SignatureRequest struct {
	Nonce     string `json:"nonce" validate:"required,lowercase,len=32"`
	PublicKey string `json:"public_key" validate:"required,lowercase,min=64,max=67"`
	Username  string `json:"username" validate:"required,lowercase,min=5"`
	Timestamp int64  `json:"timestamp" validate:"required,number,min=10"`
	Origin    string `json:"origin"`
}

type SignatureResponse struct {
	Signature string                   `json:"signature"`
	MaxTries  int                      `json:"max_tries"`
	Cadence   SignatureCadenceResponse `json:"cadence"`
}

type SignatureCadenceResponse struct {
	ReceivedAt string `json:"received_at"`
	CreatedAt  string `json:"created_at"`
	ExpiresAt  string `json:"expires_at"`
}
