package repoentity

import (
	"time"

	"github.com/oullin/database"
)

type APIKeyCreateSignatureFor struct {
	Key       *database.APIKey
	ExpiresAt time.Time
	Seed      []byte
	Origin    string
}

type FindSignatureFrom struct {
	Key        *database.APIKey
	Signature  []byte
	Origin     string
	ServerTime time.Time
}
