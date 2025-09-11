.PHONY:

APP_CADDY_CONFIG_PROD_FILE ?= caddy/Caddyfile.prod
APP_CADDY_CONFIG_LOCAL_FILE ?= caddy/Caddyfile.local

CADDY_MTLS_DIR = $(ROOT_PATH)/caddy/mtls

caddy-gen-cert:
	openssl genrsa -out $(CADDY_MTLS_DIR)/ca.key 4096
	openssl req -x509 -new -nodes -key $(CADDY_MTLS_DIR)/ca.key -sha256 -days 3650 -subj "/CN=oullin-mtls-ca" -out $(CADDY_MTLS_DIR)/ca.pem
	chmod 600 $(CADDY_MTLS_DIR)/ca.key
	chmod 644 $(CADDY_MTLS_DIR)/ca.pem

