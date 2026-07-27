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

func TestProdCaddyfileHandlesBrowserSignatureRelayAtEdge(t *testing.T) {
	caddyfile := stripCaddyComments(readProdCaddyfile(t))

	relayContract := regexp.MustCompile(`(?s)` +
		`@relay_signature_cors\s+path\s+/relay/generate-signature\*.*?` +
		`header\s+@relay_signature_cors\s+Access-Control-Allow-Origin\s+"https://oullin\.io".*?` +
		`header\s+@relay_signature_cors\s+Access-Control-Allow-Methods\s+"POST, OPTIONS".*?` +
		`header\s+@relay_signature_cors\s+Access-Control-Allow-Headers\s+"X-API-Key, X-API-Username, X-API-Signature, X-API-Timestamp, X-API-Nonce, X-Request-ID, Content-Type, User-Agent, If-None-Match, X-API-Intended-Origin".*?` +
		`@relay_signature_preflight\s*{\s*` +
		`path\s+/relay/generate-signature\*\s+` +
		`method\s+OPTIONS\s*` +
		`}.*?` +
		`handle\s+@relay_signature_preflight\s*{\s*` +
		`header\s+Access-Control-Max-Age\s+"86400"\s+` +
		`respond\s+204\s*` +
		`}.*?` +
		`@relay_signature_post\s*{\s*` +
		`path\s+/relay/generate-signature\*\s+` +
		`method\s+POST\s*` +
		`}.*?` +
		`handle\s+@relay_signature_post\s*{\s*` +
		`uri\s+strip_prefix\s+/relay\s+` +
		`reverse_proxy\s+api:8080\s*{.*?` +
		`header_up\s+X-API-Intended-Origin\s+\{http\.request\.header\.X-API-Intended-Origin\}.*?` +
		`}`)

	if !relayContract.MatchString(caddyfile) {
		t.Fatal("expected /relay/generate-signature* to be handled by the public edge and proxied to api:8080")
	}

	if !regexp.MustCompile(`handle\s+/relay/generate-signature\*\s*{\s*respond\s+405\s*}`).MatchString(caddyfile) {
		t.Fatal("expected unsupported relay signature methods to respond with 405")
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

func TestProdCaddyfileRedirectsWritingArchiveBeforeDefaultProxy(t *testing.T) {
	caddyfile := stripCaddyComments(readProdCaddyfile(t))
	oullinBlock, ok := caddyBlock(caddyfile, "oullin.io")
	if !ok {
		t.Fatal("expected production Caddyfile to contain the oullin.io site")
	}

	redirects := []struct {
		source string
		target string
	}{
		{"/post/2026-02-19-local-first-skills-platform-for-ai-agents", "https://writing.gocanto.sh/local-first-skills-platform-for-ai-agents"},
		{"/post/2023-06-21-bugs-are-inevitable-in-software-development", "https://writing.gocanto.sh/bugs-are-inevitable-in-software-development"},
		{"/post/2025-09-25-shipping-seo-for-a-single-page-app-the-pragmatic-way", "https://writing.gocanto.sh/shipping-seo-for-a-single-page-app-the-pragmatic-way"},
		{"/post/2025-10-09-when-a-real-time-feature-store-is-the-wrong-fix-for-fraud", "https://writing.gocanto.sh/when-a-real-time-feature-store-is-the-wrong-fix-for-fraud"},
		{"/post/2025-11-04-your-ai-model-is-not-your-product", "https://writing.gocanto.sh/your-ai-model-is-not-your-product"},
		{"/post/2026-04-05-go-meets-the-monolith-renaissance", "https://writing.gocanto.sh/go-meets-the-monolith-renaissance"},
		{"/post/2025-12-29-nostalgia-is-a-security-risk", "https://writing.gocanto.sh/nostalgia-is-a-security-risk"},
		{"/post/2025-11-21-demystifying-the-go-maps-engine", "https://writing.gocanto.sh/demystifying-the-go-maps-engine"},
		{"/post/2025-12-15-money-in-Go-done-properly", "https://writing.gocanto.sh/money-in-go-done-properly"},
		{"/post/2026-03-13-taming-complex-state-in-go", "https://writing.gocanto.sh/taming-complex-state-in-go"},
		{"/post/2026-03-31-laravel-collections-to-go", "https://writing.gocanto.sh/laravel-collections-to-go"},
		{"/post/2025-10-29-turbocharging-web-performance-automated-early-hints", "https://writing.gocanto.sh/turbocharging-web-performance-automated-early-hints"},
		{"/post/2025-11-13-shipping-observability-for-oullin-infrastructure", "https://writing.gocanto.sh/shipping-observability-for-oullin-infrastructure"},
		{"/post/2025-11-26-the-case-of-the-mismatched-clocks-lesson-secure-timing", "https://writing.gocanto.sh/the-case-of-the-mismatched-clocks-lesson-secure-timing"},
		{"/post/2026-02-26-stop-installing-python-just-to-convert-files-to-markdown", "https://writing.gocanto.sh/stop-installing-python-just-to-convert-files-to-markdown"},
		{"/post/2025-11-15-debugging-multi-layered-docker-deployment", "https://writing.gocanto.sh/debugging-multi-layered-docker-deployment"},
		{"/post/2023-06-22-positive-emotions-reduce-negative-emotions", "https://writing.gocanto.sh/positive-emotions-reduce-negative-emotions"},
		{"/post/2025-04-02-embrace-growth-through-movement", "https://writing.gocanto.sh/embrace-growth-through-movement"},
		{"/post/2025-10-17-the-operating-system-of-you-an-engineering-leader-reflection", "https://writing.gocanto.sh/the-operating-system-of-you-an-engineering-leader-reflection"},
		{"/post/2025-11-03-honesty-establishes-trust-enabling-process-formation-progression", "https://writing.gocanto.sh/honesty-establishes-trust-enabling-process-formation-progression"},
		{"/post/2025-11-30-work-is-not-family-and-that-is-a-good-thing", "https://writing.gocanto.sh/work-is-not-family-and-that-is-a-good-thing"},
		{"/post/2026-02-24-engineering-management-mistakes-we-need-to-stop-making.md", "https://writing.gocanto.sh/engineering-management-mistakes-we-need-to-stop-making"},
		{"/post/2025-12-05-handling-money-in-php", "https://writing.gocanto.sh/handling-money-in-php"},
		{"/post/2025-10-15-building-oullin", "https://writing.gocanto.sh/building-oullin"},
	}

	defaultProxy := strings.Index(oullinBlock, "reverse_proxy web:80")
	if defaultProxy == -1 {
		t.Fatal("expected oullin.io site to contain the default web proxy")
	}

	for _, redirect := range redirects {
		t.Run(redirect.source, func(t *testing.T) {
			directive := "redir " + redirect.source + " " + redirect.target + " 301"
			position := strings.Index(oullinBlock, directive)
			if position == -1 {
				t.Fatalf("expected exact writing redirect %q", directive)
			}
			if position > defaultProxy {
				t.Fatalf("expected %s redirect before the default web proxy", redirect.source)
			}
		})
	}

	for _, directive := range []string{
		"redir /writing https://writing.gocanto.sh/ 301",
		"redir /writing/ https://writing.gocanto.sh/ 301",
		"redir /tags/* https://writing.gocanto.sh/ 301",
	} {
		position := strings.Index(oullinBlock, directive)
		if position == -1 {
			t.Fatalf("expected archive redirect %q", directive)
		}
		if position > defaultProxy {
			t.Fatalf("expected archive redirect %q before the default web proxy", directive)
		}
	}

	if strings.Contains(oullinBlock, "redir /post/*") {
		t.Fatal("must not use a broad /post/* redirect")
	}
}
