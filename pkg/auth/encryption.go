package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

func GenerateAESKey() ([]byte, error) {
	key := make([]byte, EncryptionKeyLength)

	if _, err := rand.Read(key); err != nil {
		return []byte(""), err
	}

	return key, nil
}

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// GCM is an authenticated encryption mode that provides confidentiality and integrity.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// A nonce is a "number used once" to ensure that the same plaintext
	// encrypts to different ciphertexts each time. It must be unique for each encryption.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal will Encrypt the data and append the authentication tag.
	// We prepend the nonce to the ciphertext for use during decryption.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// The nonce is prepended to the ciphertext.
	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Open will decrypt and authenticate the message.
	plaintext, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func CreateSignatureFrom(message, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))

	return hex.EncodeToString(mac.Sum(nil))
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
