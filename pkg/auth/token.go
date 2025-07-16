package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func SetupNewAccount(accountName string, TokenLength int) (*Token, error) {
	token := Token{}

	pk, err := generateSecureToken(TokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	sk, err := generateSecureToken(TokenLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	token.PublicKey = pk
	token.SecretKey = sk
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

func ValidateBearerToken(seed string) (*ValidatedToken, error) {
	validated := ValidatedToken{
		Token: strings.TrimSpace(seed),
	}

	if validated.Token == "" {
		return nil, fmt.Errorf("token not found or invalid")
	}

	if strings.HasPrefix(validated.Token, PublicKeyPrefix) {
		validated.AuthLevel = strings.TrimSpace(LevelPublic)
		return &validated, nil
	}

	if strings.HasPrefix(validated.Token, SecretKeyPrefix) {
		validated.AuthLevel = strings.TrimSpace(LevelSecret)
		return &validated, nil
	}

	return nil, fmt.Errorf("the given token [%s] is not valid", validated.Token)
}
