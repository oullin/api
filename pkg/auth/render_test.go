package auth

import "testing"

func TestSafeDisplay(t *testing.T) {
	tok := "sk_1234567890123456abcd"
	d := SafeDisplay(tok)
	expected := "sk_1234567890..."

	if d != expected {
		t.Fatalf("expected %s got %s", expected, d)
	}
}

func TestSafeDisplayShort(t *testing.T) {
	tok := "pk_short"

	if SafeDisplay(tok) != tok {
		t.Fatalf("expected same")
	}
}
