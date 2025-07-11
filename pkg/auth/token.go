package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type Token struct {
	Public  string `validate:"required,min=10"`
	Private string `validate:"required,min=10"`
}

func (t Token) IsInvalid(seed string) bool {
	return !t.IsValid(seed)
}

func (t Token) IsValid(seed string) bool {
	token := strings.TrimSpace(t.Public)
	salt := strings.TrimSpace(t.Private)
	externalSalt := strings.TrimSpace(seed)

	if salt != externalSalt {
		return false
	}

	hash := sha256.New()
	hash.Write([]byte(externalSalt))
	bytes := hash.Sum(hash.Sum(nil))

	encodeToString := strings.TrimSpace(
		hex.EncodeToString(bytes),
	)

	return token == encodeToString
}
