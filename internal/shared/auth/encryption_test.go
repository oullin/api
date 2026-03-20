package auth_test

import (
	"testing"

	"github.com/oullin/internal/shared/auth"
)

func TestEncryptDecrypt(t *testing.T) {
	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate aes key: %v", err)
	}

	if len(key) != auth.EncryptionKeyLength {
		t.Fatalf("expected key length %d, got %d", auth.EncryptionKeyLength, len(key))
	}

	plain := []byte("hello")
	enc, err := auth.Encrypt(plain, key)

	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	dec, err := auth.Decrypt(enc, key)

	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if string(dec) != "hello" {
		t.Fatalf("expected decrypted text 'hello', got %q", string(dec))
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate first key: %v", err)
	}

	other, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate second key: %v", err)
	}

	enc, err := auth.Encrypt([]byte("hello"), key)

	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := auth.Decrypt(enc, other); err == nil {
		t.Fatalf("expected decryption error with wrong key")
	}
}

func TestDecryptShortCipher(t *testing.T) {
	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	if _, err := auth.Decrypt([]byte("short"), key); err == nil {
		t.Fatalf("expected error for short cipher text")
	}
}

func TestValidateTokenFormatEmpty(t *testing.T) {
	if auth.ValidateTokenFormat(" ") == nil {
		t.Fatalf("expected validation error for empty token")
	}
}

func TestCreateSignatureFrom(t *testing.T) {
	sig1 := auth.CreateSignatureFrom("msg", "secret")
	sig2 := auth.CreateSignatureFrom("msg", "secret")

	if sig1 != sig2 {
		t.Fatalf("expected identical signatures for same message and secret")
	}
}

func TestValidateTokenFormat(t *testing.T) {
	if auth.ValidateTokenFormat("pk_1234567890123") != nil {
		t.Fatalf("expected valid token format to pass validation")
	}

	if auth.ValidateTokenFormat("bad") == nil {
		t.Fatalf("expected invalid token format to fail validation")
	}
}
