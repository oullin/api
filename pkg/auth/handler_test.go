package auth

import "testing"

func TestTokenHandlerLifecycle(t *testing.T) {
	key, err := GenerateAESKey()

	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	h, err := MakeTokensHandler(key)

	if err != nil {
		t.Fatalf("make handler: %v", err)
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
		t.Fatalf("decode mismatch")
	}
}

func TestMakeTokensHandlerError(t *testing.T) {
	_, err := MakeTokensHandler([]byte("short"))

	if err == nil {
		t.Fatalf("expected error for short key")
	}
}

func TestSetupNewAccountErrors(t *testing.T) {
	key, err := GenerateAESKey()

	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	h, err := MakeTokensHandler(key)

	if err != nil {
		t.Fatalf("make handler: %v", err)
	}

	if _, err := h.SetupNewAccount("ab"); err == nil {
		t.Fatalf("expected error for short name")
	}

	badHandler := &TokenHandler{EncryptionKey: []byte("short")}
	if _, err := badHandler.SetupNewAccount("tester"); err == nil {
		t.Fatalf("expected encrypt error")
	}
}

func TestDecodeTokensForError(t *testing.T) {
	key, err := GenerateAESKey()

	if err != nil {
		t.Fatalf("key err: %v", err)
	}
	h, err := MakeTokensHandler(key)

	if err != nil {
		t.Fatalf("make handler: %v", err)
	}
	if _, err := h.DecodeTokensFor("acc", []byte("bad"), []byte("bad")); err == nil {
		t.Fatalf("expected error")
	}
}
