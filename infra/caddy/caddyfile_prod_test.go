package caddy_test

import (
	"os"
	"regexp"
	"testing"
)

func readProdCaddyfile(t *testing.T) string {
	t.Helper()

	content, err := os.ReadFile("Caddyfile.prod")
	if err != nil {
		t.Fatalf("read production Caddyfile: %v", err)
	}

	return string(content)
}

func TestProdCaddyfileBlocksPublicSignatureEndpoint(t *testing.T) {
	caddyfile := readProdCaddyfile(t)

	if !regexp.MustCompile(`@protected_public\s+path\s+/api/generate-signature\*\s+/api/metrics`).MatchString(caddyfile) {
		t.Fatal("expected public protected matcher to include /api/generate-signature*")
	}

	if !regexp.MustCompile(`handle\s+@protected_public\s*{\s*respond\s+403\s*}`).MatchString(caddyfile) {
		t.Fatal("expected public protected matcher to respond with 403")
	}
}

func TestProdCaddyfileKeepsSignatureEndpointBehindMTLS(t *testing.T) {
	caddyfile := readProdCaddyfile(t)

	required := []string{
		":8443 {",
		"tls /etc/caddy/mtls/server.pem /etc/caddy/mtls/server.key",
		"mode require_and_verify",
		"trust_pool file /etc/caddy/mtls/ca.pem",
		"@sig path /api/generate-signature*",
		"uri strip_prefix /api",
		"reverse_proxy api:8080",
	}

	for _, item := range required {
		if !regexp.MustCompile(regexp.QuoteMeta(item)).MatchString(caddyfile) {
			t.Fatalf("expected production Caddyfile to contain %q", item)
		}
	}
}
