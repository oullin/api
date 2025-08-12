package auth

import (
	"testing"
	"time"
)

func TestJWTHandlerGenerateValidate(t *testing.T) {
	h, err := MakeJWTHandler([]byte("supersecretkey123"), time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	token, err := h.Generate("alice")
	if err != nil {
		t.Fatalf("generate token err: %v", err)
	}

	claims, err := h.Validate(token)
	if err != nil {
		t.Fatalf("validate token err: %v", err)
	}

	if claims.Username != "alice" {
		t.Fatalf("expected alice got %s", claims.Username)
	}
}

func TestJWTHandlerValidateFail(t *testing.T) {
	h, err := MakeJWTHandler([]byte("anothersecretkey"), time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	if _, err := h.Validate("invalid.token"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
