package auth

import "testing"

func TestSafeDisplay(t *testing.T) {
	tok := "sk_1234567890123456abcd"
	d := SafeDisplay(tok)

	if d == tok || len(d) >= len(tok) {
		t.Fatalf("masking failed")
	}
}
