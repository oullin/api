.PHONY: caddy-gen-certs caddy-del-certs caddy-validate

CADDY_MTLS_DIR = $(ROOT_PATH)/caddy/mtls
APP_CADDY_CONFIG_PROD_FILE ?= caddy/Caddyfile.prod
APP_CADDY_CONFIG_LOCAL_FILE ?= caddy/Caddyfile.local

caddy-gen-certs:
	@set -eu; \
	mkdir -p "$(CADDY_MTLS_DIR)"; chmod 700 "$(CADDY_MTLS_DIR)"; \
	if [ -d "$(CADDY_MTLS_DIR)/ca.pem" ]; then \
	  printf "$(RED)✘ ERROR:$(NC) %s is a directory. Move or remove it.\n" "$(CADDY_MTLS_DIR)/ca.pem"; \
	fi; \
	if [ -e "$(CADDY_MTLS_DIR)/ca.key" ] || [ -e "$(CADDY_MTLS_DIR)/ca.pem" ]; then \
	  printf "$(YELLOW)⚠️  CA already exists in %s.$(NC)\n" "$(CADDY_MTLS_DIR)"; \
	  printf "$(CYAN)👉 Remove it with 'make caddy-clean-certs' if you want to recreate.$(NC)\n"; \
	else \
	  umask 077; \
	  printf "$(BLUE)🔑 Generating CA private key...$(NC)\n"; \
	  openssl genrsa -out "$(CADDY_MTLS_DIR)/ca.key" 4096 >/dev/null 2>&1; \
	  printf "$(BLUE)📜 Creating self-signed CA certificate...$(NC)\n"; \
	  openssl req -x509 -new -key "$(CADDY_MTLS_DIR)/ca.key" -sha256 -days 3650 \
	    -subj "/CN=oullin-mtls-ca" -out "$(CADDY_MTLS_DIR)/ca.pem" >/dev/null 2>&1; \
	  printf '01\n' > "$(CADDY_MTLS_DIR)/ca.srl"; \
	  chmod 600 "$(CADDY_MTLS_DIR)/ca.key"; \
	  chmod 644 "$(CADDY_MTLS_DIR)/ca.pem" "$(CADDY_MTLS_DIR)/ca.srl"; \
	  printf "$(GREEN)✅ CA written to %s$(NC)\n" "$(CADDY_MTLS_DIR)"; \
	  printf "$(WHITE)🔍 CA fingerprint:$(NC)\n"; \
	  openssl x509 -noout -fingerprint -sha256 -in "$(CADDY_MTLS_DIR)/ca.pem" | sed 's/^/   /'; \
	fi

caddy-del-certs:
	@set -eu; \
	rm -f "$(CADDY_MTLS_DIR)/ca.key" "$(CADDY_MTLS_DIR)/ca.pem" "$(CADDY_MTLS_DIR)/ca.srl"; \
	printf "$(BLUE)✅ files removed from [$(NC)$(CADDY_MTLS_DIR)$(BLUE)]$(NC)\n"

caddy-validate:
	docker run --rm \
      -v "$(ROOT_PATH)/caddy/Caddyfile.prod:/etc/caddy/Caddyfile:ro" \
      -v "$(ROOT_PATH)/caddy/mtls:/etc/caddy/mtls:ro" \
      caddy:2.10.0 caddy validate --config /etc/caddy/Caddyfile
