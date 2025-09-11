# [2025-09-11] Postmortem — Web/API Containers mTLS Relay

## Actual architecture
- Containers on one VPS
	- [web](https://github.com/oullin/web) Caddy serves the built Vue app on :80 and exposes a relay endpoint.
	- api Caddy terminates public TLS on :80 and :443 and exposes an internal listener on :8443 that requires client certs.
- Network
	- Both Caddies are on the shared Docker bridge caddy_net. The internal hop is web → api:8443 inside this network.
- mTLS model
	- A private CA issues client certificates.
	- api Caddy trusts the CA via trust_pool.
	- web Caddy holds client.pem and client.key and presents them when proxying to api:8443.
- Paths
	- Protected API path: /api/generate-signature.
	- Frontend calls GET /relay/generate-signature* on web Caddy which rewrites and proxies to the protected API path over mTLS.

## The issue
- Generating the web client key failed:
	- genrsa: Can't open '.../caddy/mtls/client.key' for writing, Is a directory.
- Earlier runs also failed to sign the CSR because the CA path pointed to a non-existent location and ENV_API_LOCAL_DIR was not exported into the recipe shell.

## Analysis
- [api] File vs. directory collision
	- A directory was created at the path where a file should exist. OpenSSL could not write client.key over a directory.
- [web] CA location and trust
	- The client cert must be signed by a CA that the API trust pool contains. Using a different repo path is fine as long as the same CA public pem is mounted into the API container.
- [web] Make shell env
	- Make included .env but did not export it, so ENV_API_LOCAL_DIR was empty inside the recipe. The recipe therefore could not find ca.pem and ca.key.
- Repeatability
	- Missing guards for directory collisions and for absent CA files made the process brittle.

## The solution
1) Clean and regenerate client materials in the web repo
	- Remove accidental directory at caddy/mtls/client.key
	- Ensure parent dir exists (mkdir -p)
	- Export ENV_API_LOCAL_DIR pointing to the API CA directory

2) Harden the Makefile target for the web repo
	- Validate prerequisites (CA files exist)
	- Refuse to proceed if client.key or client.pem are directories
	- Use umask 077 and lock down permissions
	- Clean CSR after signing

3) Web Caddy relay
	- Expose /relay/generate-signature
	- reverse_proxy to https://oullin_proxy_prod:8443
	- Present client_certificate /etc/caddy/mtls/client.pem /etc/caddy/mtls/client.key
	- Map relay path to /api/generate-signature
	- Optionally set tls_insecure_skip_verify or mount the API root CA and use root_ca

4) Web compose mounts
	- Mount client.pem and client.key read-only into /etc/caddy/mtls/

5) Validation
	- Rebuild and start the web Caddy
	- From inside the web Caddy container, curl http://127.0.0.1/relay/generate-signature?ping=1 (expect 200)
	- Public path https://oullin.io/api/generate-signature should respond 403

## Outro
The web container failed because of a file system mismatch and missing environment exports. By cleaning the accidental 
directory, hardening the Makefile to validate inputs and prevent collisions, and confirming the mTLS trust chain, 
the relay now works reliably. This pattern keeps the CA private, reduces risk, and makes the web to API hop explicit 
and testable.
