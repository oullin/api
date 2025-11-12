.PHONY: caddy-gen-certs caddy-del-certs caddy-validate caddy-fresh caddy-restart

CADDY_MTLS_DIR = $(ROOT_PATH)/infra/caddy/mtls
APP_CADDY_CONFIG_PROD_FILE ?= infra/caddy/Caddyfile.prod
APP_CADDY_CONFIG_LOCAL_FILE ?= infra/caddy/Caddyfile.local

caddy-restart:
	docker compose up -d --force-recreate caddy_prod

caddy-fresh:
	@make caddy-del-certs
	@make caddy-gen-certs

caddy-gen-certs:
	@set -eu; \
	mkdir -p "$(CADDY_MTLS_DIR)"; chmod 700 "$(CADDY_MTLS_DIR)"; \
	if [ -d "$(CADDY_MTLS_DIR)/ca.pem" ]; then \
	  printf "$(RED)✘ ERROR:$(NC) %s is a directory. Move or remove it.\n" "$(CADDY_MTLS_DIR)/ca.pem"; \
	  exit 1; \
	fi; \
	if [ -e "$(CADDY_MTLS_DIR)/ca.key" ] || [ -e "$(CADDY_MTLS_DIR)/ca.pem" ]; then \
	  printf "$(YELLOW)⚠️  CA already exists in %s.$(NC)\n" "$(CADDY_MTLS_DIR)"; \
	  printf "$(CYAN)👉 Remove it with 'make caddy-del-certs' if you want to recreate.$(NC)\n"; \
	else \
	  umask 077; \
	  printf "$(BLUE)🔑 Generating CA private key...$(NC)\n"; \
	  openssl genrsa -out "$(CADDY_MTLS_DIR)/ca.key" 4096; \
	  printf "$(BLUE)📜 Creating self-signed CA certificate...$(NC)\n"; \
	  openssl req -x509 -new -key "$(CADDY_MTLS_DIR)/ca.key" -sha256 -days 3650 \
	    -subj "/CN=oullin-mtls-ca" -out "$(CADDY_MTLS_DIR)/ca.pem"; \
	  printf '01\n' > "$(CADDY_MTLS_DIR)/ca.srl"; \
	  chmod 600 "$(CADDY_MTLS_DIR)/ca.key"; \
	  chmod 644 "$(CADDY_MTLS_DIR)/ca.pem" "$(CADDY_MTLS_DIR)/ca.srl"; \
	  printf "$(GREEN)✅ CA written to %s$(NC)\n" "$(CADDY_MTLS_DIR)"; \
	  printf "$(WHITE)🔍 CA fingerprint:$(NC)\n"; \
	  openssl x509 -noout -fingerprint -sha256 -in "$(CADDY_MTLS_DIR)/ca.pem" | sed 's/^/   /'; \
	fi; \
	if [ -e "$(CADDY_MTLS_DIR)/server.key" ] || [ -e "$(CADDY_MTLS_DIR)/server.pem" ]; then \
	  printf "$(YELLOW)⚠️  Server certificate already exists.$(NC)\n"; \
	else \
	  umask 077; \
	  printf "$(BLUE)🔑 Generating Server private key...$(NC)\n"; \
	  openssl genrsa -out "$(CADDY_MTLS_DIR)/server.key" 4096; \
	  printf "$(WHITE)📜 Creating Server signing request...$(NC)\n"; \
	  openssl req -new -key "$(CADDY_MTLS_DIR)/server.key" -subj "/CN=oullin_proxy_prod" -out "$(CADDY_MTLS_DIR)/server.csr"; \
	  printf "$(WHITE)📝 Creating SAN extension file...$(NC)\n"; \
	  EXTFILE="$$(mktemp)"; trap 'rm -f "$$EXTFILE"' EXIT; \
	  printf "subjectAltName=DNS:oullin_proxy_prod,DNS:localhost,IP:127.0.0.1\nextendedKeyUsage=serverAuth\n" > "$$EXTFILE"; \
	  printf "$(NC)🖊️ Signing Server certificate with CA (including SANs)...$(NC)\n"; \
	  openssl x509 -req -in "$(CADDY_MTLS_DIR)/server.csr" \
	    -CA "$(CADDY_MTLS_DIR)/ca.pem" -CAkey "$(CADDY_MTLS_DIR)/ca.key" -CAserial "$(CADDY_MTLS_DIR)/ca.srl" \
	    -out "$(CADDY_MTLS_DIR)/server.pem" -days 1095 -sha256 -extfile "$$EXTFILE"; \
	  rm -f "$$EXTFILE"; \
	  chmod 600 "$(CADDY_MTLS_DIR)/server.key"; \
	  chmod 644 "$(CADDY_MTLS_DIR)/server.pem"; \
	  rm "$(CADDY_MTLS_DIR)/server.csr"; \
	  printf "\n$(GREEN)✅ Server certificate written to %s$(NC)\n" "$(CADDY_MTLS_DIR)"; \
	fi; \
	printf "\n$(CYAN)🔎 Verifying server cert against CA ...$(NC)\n"; \
	openssl verify -CAfile "$(CADDY_MTLS_DIR)/ca.pem" "$(CADDY_MTLS_DIR)/server.pem"; \
	printf "\n$(BLUE)🏗️ Verifying SAN ...$(NC)\n"; \
	openssl x509 -in infra/caddy/mtls/server.pem -text -noout | grep -A 2 "Subject Alternative Name"

caddy-del-certs:
	@set -eu; \
	rm -f "$(CADDY_MTLS_DIR)/ca.key" \
	      "$(CADDY_MTLS_DIR)/ca.pem" \
	      "$(CADDY_MTLS_DIR)/ca.srl" \
	      "$(CADDY_MTLS_DIR)/server.key" \
	      "$(CADDY_MTLS_DIR)/server.pem"; \
	printf "$(BLUE)✅ files removed from [$(NC)$(CADDY_MTLS_DIR)$(BLUE)]$(NC)\n"

caddy-validate:
	@docker run --rm \
	  -v "$(ROOT_PATH)/infra/caddy/Caddyfile.prod:/etc/caddy/Caddyfile:ro" \
	  -v "$(ROOT_PATH)/infra/caddy/mtls:/etc/caddy/mtls:ro" \
	  caddy:2.10.0 caddy validate --config /etc/caddy/Caddyfile
