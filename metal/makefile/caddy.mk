.PHONY: caddy-gen-certs caddy-clean-certs caddy-validate

CADDY_MTLS_DIR = $(ROOT_PATH)/caddy/mtls
APP_CADDY_CONFIG_PROD_FILE ?= caddy/Caddyfile.prod
APP_CADDY_CONFIG_LOCAL_FILE ?= caddy/Caddyfile.local

caddy-gen-certs:
	@set -eu; \
	mkdir -p "$(CADDY_MTLS_DIR)"; \
	# fail fast if someone accidentally created a folder named ca.pem
	if [ -d "$(CADDY_MTLS_DIR)/ca.pem" ]; then \
	  echo "ERROR: $(CADDY_MTLS_DIR)/ca.pem is a directory. Move or remove it."; exit 1; \
	fi; \
	# refuse to overwrite an existing CA
	if [ -e "$(CADDY_MTLS_DIR)/ca.key" ] || [ -e "$(CADDY_MTLS_DIR)/ca.pem" ]; then \
	  echo "CA already exists in $(CADDY_MTLS_DIR). Remove it or run 'make caddy-clean-ca' first."; exit 1; \
	fi; \
	umask 077; \
	openssl genrsa -out "$(CADDY_MTLS_DIR)/ca.key" 4096; \
	openssl req -x509 -new -key "$(CADDY_MTLS_DIR)/ca.key" -sha256 -days 3650 \
		-subj "/CN=oullin-mtls-ca" -out "$(CADDY_MTLS_DIR)/ca.pem"; \
	# initialize serial file once
	printf '01\n' > "$(CADDY_MTLS_DIR)/ca.srl"; \
	chmod 600 "$(CADDY_MTLS_DIR)/ca.key"; \
	chmod 644 "$(CADDY_MTLS_DIR)/ca.pem" "$(CADDY_MTLS_DIR)/ca.srl"; \
	echo "✓ CA written to $(CADDY_MTLS_DIR)"

caddy-clean-certs:
	@set -eu; \
	rm -f "$(CADDY_MTLS_DIR)/ca.key" "$(CADDY_MTLS_DIR)/ca.pem" "$(CADDY_MTLS_DIR)/ca.srl"; \
	echo "✓ Removed CA files from $(CADDY_MTLS_DIR)"

# --- Mac:
#     Needs to be locally installed: https://formulae.brew.sh/formula/caddy
caddy-validate:
	caddy fmt --overwrite $(APP_CADDY_CONFIG_PROD_FILE)
	caddy validate --config $(APP_CADDY_CONFIG_PROD_FILE)
	caddy fmt --overwrite $(APP_CADDY_CONFIG_LOCAL_FILE)
	caddy validate --config $(APP_CADDY_CONFIG_LOCAL_FILE)
