package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func SetupNewAccount(accountName string, TokenLength int) (*Token, error) {
	token := Token{}

	if len(accountName) < AccountNameMinLength {
		return nil, fmt.Errorf("account name must be at least %d characters", AccountNameMinLength)
	}

	pk, err := generateSecureToken(TokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	sk, err := generateSecureToken(TokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	token.PublicKey = PublicKeyPrefix + pk
	token.SecretKey = SecretKeyPrefix + sk
	token.Length = TokenLength
	token.AccountName = accountName

	return &token, nil
}

func generateSecureToken(length int) (string, error) {
	if length < TokenMinLength {
		return "", fmt.Errorf("the token length should be >= %d", length)
	}

	salt := make([]byte, length)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate secure tokens salt: %v", err)
	}

	hasher := sha256.New()
	hasher.Write(salt)

	// Get the resulting hash and encode it as a hex string.
	hashBytes := hasher.Sum(nil)

	return hex.EncodeToString(hashBytes), nil
}

func ValidateTokenFormat(seed string) error {
	token := strings.TrimSpace(seed)

	if token == "" || len(token) < TokenMinLength {
		return fmt.Errorf("token not found or invalid")
	}

	if !strings.HasPrefix(token, PublicKeyPrefix) || !strings.HasPrefix(token, SecretKeyPrefix) {
		return fmt.Errorf("invalid token prefix")
	}

	return fmt.Errorf("the given token [%s] is not valid", token)
}

func CreateSignatureFrom(message, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))

	return hex.EncodeToString(mac.Sum(nil))
}

func SafeDisplay(secret string) string {
	var prefixLen int
	visibleChars := 8

	if strings.HasPrefix(secret, PublicKeyPrefix) {
		prefixLen = len(PublicKeyPrefix)
	} else {
		prefixLen = len(SecretKeyPrefix)
	}

	if len(secret) <= prefixLen+visibleChars {
		return secret
	}

	return secret[:prefixLen+visibleChars] + "..."
}

func (t Token) HasInValidSignature(receivedSignature string) bool {
	return !t.HasValidSignature(receivedSignature)
}

func (t Token) HasValidSignature(receivedSignature string) bool {
	signature := CreateSignatureFrom(
		t.AccountName,
		t.SecretKey,
	)

	return hmac.Equal(
		[]byte(signature),
		[]byte(receivedSignature),
	)
}
