package caddy_test

import (
	"os"
	"regexp"
	"strings"
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

func protectedPublicPaths(caddyfile string) map[string]bool {
	paths := make(map[string]bool)

	for _, line := range strings.Split(caddyfile, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != "@protected_public" || fields[1] != "path" {
			continue
		}

		for _, path := range fields[2:] {
			paths[path] = true
		}

		return paths
	}

	return paths
}

func stripCaddyComments(caddyfile string) string {
	var lines []string

	for _, line := range strings.Split(caddyfile, "\n") {
		beforeComment, _, _ := strings.Cut(line, "#")
		lines = append(lines, beforeComment)
	}

	return strings.Join(lines, "\n")
}

func caddyBlock(caddyfile, name string) (string, bool) {
	lines := strings.Split(caddyfile, "\n")
	start := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == name+" {" {
			start = i
			break
		}
	}

	if start == -1 {
		return "", false
	}

	depth := 0
	var block []string
	for _, line := range lines[start:] {
		block = append(block, line)
		depth += strings.Count(line, "{")
		depth -= strings.Count(line, "}")

		if depth == 0 {
			return strings.Join(block, "\n"), true
		}
	}

	return "", false
}

func TestProdCaddyfileBlocksPublicSignatureEndpoint(t *testing.T) {
	caddyfile := readProdCaddyfile(t)

	protectedPaths := protectedPublicPaths(caddyfile)
	for _, requiredPath := range []string{"/api/generate-signature*", "/api/metrics"} {
		if !protectedPaths[requiredPath] {
			t.Fatalf("expected public protected matcher to include %s", requiredPath)
		}
	}

	if !regexp.MustCompile(`handle\s+@protected_public\s*{\s*respond\s+403\s*}`).MatchString(caddyfile) {
		t.Fatal("expected public protected matcher to respond with 403")
	}
}

func TestProdCaddyfileKeepsSignatureEndpointBehindMTLS(t *testing.T) {
	caddyfile := readProdCaddyfile(t)
	mtlsBlock, ok := caddyBlock(stripCaddyComments(caddyfile), ":8443")
	if !ok {
		t.Fatal("expected production Caddyfile to contain the :8443 mTLS listener")
	}

	mtlsSignatureBlock := regexp.MustCompile(`(?s):8443\s*{\s*` +
		`tls\s+/etc/caddy/mtls/server\.pem\s+/etc/caddy/mtls/server\.key\s*{.*?` +
		`mode\s+require_and_verify\s+` +
		`trust_pool\s+file\s+/etc/caddy/mtls/ca\.pem.*?` +
		`@sig\s+path\s+/api/generate-signature\*.*?` +
		`handle\s+@sig\s*{\s*` +
		`uri\s+strip_prefix\s+/api\s+` +
		`reverse_proxy\s+api:8080\s*` +
		`}`)

	if !mtlsSignatureBlock.MatchString(mtlsBlock) {
		t.Fatal("expected /api/generate-signature* to be handled by api:8080 inside the :8443 mTLS listener")
	}
}
