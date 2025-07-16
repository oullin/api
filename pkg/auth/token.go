package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type TokenHandler struct {
	EncryptionKey        []byte
	TokenMinLength       int
	AccountNameMinLength int
}

func MakeTokenHandler(encryptionKey []byte, accountNameMinLength, tokenMinLength int) (*TokenHandler, error) {
	if tokenMinLength < TokenMinLength {
		return nil, fmt.Errorf("the token length should be at least %d", TokenMinLength)
	}

	if accountNameMinLength < AccountNameMinLength {
		return nil, fmt.Errorf("the token length should be at least %d", AccountNameMinLength)
	}

	return &TokenHandler{
		EncryptionKey:        encryptionKey,
		TokenMinLength:       tokenMinLength,
		AccountNameMinLength: accountNameMinLength,
	}, nil
}

func (t *TokenHandler) SetupNewAccount(accountName string) (*Token, error) {
	token := Token{}

	if len(accountName) < AccountNameMinLength {
		return nil, fmt.Errorf("account name must be at least %d characters", AccountNameMinLength)
	}

	pk, err := t.generateSecureToken(PublicKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	sk, err := t.generateSecureToken(SecretKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	token.AccountName = accountName
	token.PublicKey = pk.PlainText
	token.EncryptedPublicKey = pk.EncryptedText
	token.SecretKey = sk.PlainText
	token.EncryptedSecretKey = sk.EncryptedText

	return &token, nil
}

func (t *TokenHandler) generateSecureToken(prefix string) (*SecureToken, error) {
	salt := make([]byte, t.TokenMinLength)

	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate secure tokens salt: %v", err)
	}

	hasher := sha256.New()
	hasher.Write(salt)

	// Get the resulting hash and encode it as a hex string.
	hashBytes := hasher.Sum(nil)

	text := prefix + hex.EncodeToString(hashBytes)
	encryptedText, err := Encrypt([]byte(text), t.EncryptionKey)

	if err != nil {
		return nil, fmt.Errorf("failed to Encrypt: %w", err)
	}

	return &SecureToken{
		PlainText:     text,
		EncryptedText: encryptedText,
	}, nil
}

func ValidateTokenFormat(seed string) error {
	token := strings.TrimSpace(seed)

	if token == "" || len(token) < TokenMinLength {
		return fmt.Errorf("token not found or invalid")
	}

	if strings.HasPrefix(token, PublicKeyPrefix) || strings.HasPrefix(token, SecretKeyPrefix) {
		return nil
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
