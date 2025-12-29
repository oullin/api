package auth_test

import (
	"testing"

	"github.com/oullin/pkg/auth"
)

func TestTokenHandlerLifecycle(t *testing.T) {
	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	h, err := auth.NewTokensHandler(key)

	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	token, err := h.SetupNewAccount("tester")

	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	decoded, err := h.DecodeTokensFor(token.AccountName, token.EncryptedSecretKey, token.EncryptedPublicKey)

	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.PublicKey != token.PublicKey || decoded.SecretKey != token.SecretKey {
		t.Fatalf("expected decoded keys to match original tokens")
	}
}

func TestNewTokensHandlerError(t *testing.T) {
	_, err := auth.NewTokensHandler([]byte("short"))

	if err == nil {
		t.Fatalf("expected error for encryption key shorter than required length")
	}
}

func TestSetupNewAccountErrors(t *testing.T) {
	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	h, err := auth.NewTokensHandler(key)

	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	if _, err := h.SetupNewAccount("ab"); err == nil {
		t.Fatalf("expected error for account name shorter than minimum length")
	}

	badHandler := &auth.TokenHandler{EncryptionKey: []byte("short")}

	if _, err := badHandler.SetupNewAccount("tester"); err == nil {
		t.Fatalf("expected encryption error with invalid key length")
	}
}

func TestDecodeTokensForError(t *testing.T) {
	key, err := auth.GenerateAESKey()

	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	h, err := auth.NewTokensHandler(key)

	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	if _, err := h.DecodeTokensFor("acc", []byte("bad"), []byte("bad")); err == nil {
		t.Fatalf("expected error when decoding invalid encrypted tokens")
	}
}
