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
