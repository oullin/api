package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

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
