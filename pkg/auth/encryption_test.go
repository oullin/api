package auth

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateAESKey()
	if err != nil {
		t.Fatalf("key err: %v", err)
	}
	if len(key) != EncryptionKeyLength {
		t.Fatalf("invalid key length %d", len(key))
	}

	plain := []byte("hello")
	enc, err := Encrypt(plain, key)
	if err != nil {
		t.Fatalf("encrypt err: %v", err)
	}
	dec, err := Decrypt(enc, key)
	if err != nil {
		t.Fatalf("decrypt err: %v", err)
	}
	if string(dec) != "hello" {
		t.Fatalf("expected hello got %s", dec)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key, _ := GenerateAESKey()
	other, _ := GenerateAESKey()
	enc, _ := Encrypt([]byte("hello"), key)

	if _, err := Decrypt(enc, other); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCreateSignatureFrom(t *testing.T) {
	sig1 := CreateSignatureFrom("msg", "secret")
	sig2 := CreateSignatureFrom("msg", "secret")

	if sig1 != sig2 {
		t.Fatalf("signature mismatch")
	}
}

func TestValidateTokenFormat(t *testing.T) {
	if ValidateTokenFormat("pk_1234567890123") != nil {
		t.Fatalf("valid token should pass")
	}
	if ValidateTokenFormat("bad") == nil {
		t.Fatalf("invalid token should fail")
	}
}
