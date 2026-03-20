package portal_test

import (
	"testing"

	"github.com/oullin/internal/shared/portal"
)

func TestPassword_NewAndValidate(t *testing.T) {
	pw, err := portal.NewPassword("secret")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pw.Is("secret") {
		t.Fatalf("password validation failed")
	}

	if pw.Is("other") {
		t.Fatalf("password should not match")
	}

	if pw.GetHash() == "" {
		t.Fatalf("hash is empty")
	}
}
