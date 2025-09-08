package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func CreateSignature(message, secretKey []byte) []byte {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write(message)

	return mac.Sum(nil)
}

func SignatureToString(signature []byte) string {
	return hex.EncodeToString(signature)
}

func VerifySignature(message, secretKey, signature []byte) bool {
	mac := hmac.New(sha256.New, secretKey)
	_, _ = mac.Write(message)
	expected := mac.Sum(nil)

	return hmac.Equal(expected, signature)
}
