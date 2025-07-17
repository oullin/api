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

func MakeTokensHandler(encryptionKey []byte, accountNameMinLength, tokenMinLength int) (*TokenHandler, error) {
	if len(encryptionKey) != EncryptionKeyLength {
		return nil, fmt.Errorf("encryption key length must be equal to %d bytes", EncryptionKeyLength)
	}

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

func (t *TokenHandler) DecodeTokensFor(accountName string, secret, public []byte) (*Token, error) {
	var err error
	var publicKey, secretKey []byte

	if publicKey, err = Decrypt(public, t.EncryptionKey); err != nil {
		return nil, fmt.Errorf("unable to decrypt public key: %w", err)
	}

	if secretKey, err = Decrypt(secret, t.EncryptionKey); err != nil {
		return nil, fmt.Errorf("unable to decrypt secret key: %w", err)
	}

	return &Token{
		AccountName:        accountName,
		PublicKey:          string(publicKey),
		EncryptedPublicKey: public,
		SecretKey:          string(secretKey),
		EncryptedSecretKey: secret,
	}, nil
}

func (t *TokenHandler) SetupNewAccount(accountName string) (*Token, error) {
	var err error
	var pk, sk *SecureToken

	if len(accountName) < AccountNameMinLength {
		return nil, fmt.Errorf("account name must be at least %d characters", AccountNameMinLength)
	}

	if pk, err = t.generateSecureToken(PublicKeyPrefix); err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	if sk, err = t.generateSecureToken(SecretKeyPrefix); err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	return &Token{
		AccountName:        accountName,
		PublicKey:          pk.PlainText,
		EncryptedPublicKey: pk.EncryptedText,
		SecretKey:          sk.PlainText,
		EncryptedSecretKey: sk.EncryptedText,
	}, nil
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
	visibleChars := 10

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
