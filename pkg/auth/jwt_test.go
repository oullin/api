package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/oullin/database"
)

// mockRepo is a simple in-memory API key repository for tests.
type mockRepo struct {
	keys map[string]*database.APIKey
}

func (m mockRepo) FindBy(accountName string) *database.APIKey {
	return m.keys[strings.ToLower(accountName)]
}

func TestJWTHandlerGenerateValidate(t *testing.T) {
	repo := mockRepo{keys: map[string]*database.APIKey{
		"alice": {AccountName: "alice", SecretKey: []byte("supersecretkey12345")},
	}}

	h, err := MakeJWTHandler(repo, time.Minute)
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

	if claims.AccountName != "alice" {
		t.Fatalf("expected alice got %s", claims.AccountName)
	}
}

func TestJWTHandlerValidateFail(t *testing.T) {
	repo := mockRepo{keys: map[string]*database.APIKey{}}
	h, err := MakeJWTHandler(repo, time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	if _, err := h.Validate("invalid.token"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
