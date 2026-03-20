package auth_test

import (
	"testing"

	"github.com/oullin/internal/shared/auth"
)

func TestSafeDisplay(t *testing.T) {
	tok := "sk_1234567890123456abcd"
	d := auth.SafeDisplay(tok)
	expected := "sk_1234567890..."

	if d != expected {
		t.Fatalf("expected %q, got %q", expected, d)
	}
}

func TestSafeDisplayShort(t *testing.T) {
	tok := "pk_short"
	d := auth.SafeDisplay(tok)

	if d != tok {
		t.Fatalf("expected short token to be displayed unchanged, got %q", d)
	}
}
