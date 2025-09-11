.PHONY: caddy-gen-cert caddy-validate

CADDY_MTLS_DIR = $(ROOT_PATH)/caddy/mtls
APP_CADDY_CONFIG_PROD_FILE ?= caddy/Caddyfile.prod
APP_CADDY_CONFIG_LOCAL_FILE ?= caddy/Caddyfile.local

caddy-gen-cert:
	openssl genrsa -out $(CADDY_MTLS_DIR)/ca.key 4096
	openssl req -x509 -new -nodes -key $(CADDY_MTLS_DIR)/ca.key -sha256 -days 3650 -subj "/CN=oullin-mtls-ca" -out $(CADDY_MTLS_DIR)/ca.pem
	openssl rand -hex 16 | tr '[:lower:]' '[:upper:]' > $(CADDY_MTLS_DIR)/ca.srl
	chmod 600 $(CADDY_MTLS_DIR)/ca.key
	chmod 644 $(CADDY_MTLS_DIR)/ca.pem
	chmod 644 $(CADDY_MTLS_DIR)/ca.srl

# --- Mac:
#     Needs to be locally installed: https://formulae.brew.sh/formula/caddy
caddy-validate:
	caddy fmt --overwrite $(APP_CADDY_CONFIG_PROD_FILE)
	caddy validate --config $(APP_CADDY_CONFIG_PROD_FILE)
	caddy fmt --overwrite $(APP_CADDY_CONFIG_LOCAL_FILE)
	caddy validate --config $(APP_CADDY_CONFIG_LOCAL_FILE)
