.PHONY: build-local build-ci build-prod build-release build-deploy build-local-restart build-prod-force build-fresh

BUILD_VERSION ?= latest
BUILD_PACKAGE_OWNER := oullin
DB_INFRA_ROOT_PATH ?= $(ROOT_PATH)/database/infra
DB_INFRA_SCRIPTS_PATH ?= $(DB_INFRA_ROOT_PATH)/scripts


build-fresh:
	make fresh && \
	make db:fresh && \
	make db:migrate && \
	make db:seed

build-local:
	docker compose --profile local up --build -d

build-local-restart:
	docker compose --profile local down && \
	docker compose --profile local up --build -d

build-ci:
	@printf "\n$(CYAN)Building production images for CI$(NC)\n"
	# This 'build' command only builds the images; it does not run them.
	@docker compose --profile prod build

# --- Deprecated
#     We should always deploy builds from the CI and not build again in servers.
build-prod:
	@DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
	DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
	DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" \
	docker compose --profile prod up --build -d

build-deploy:
	@DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
	DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
	DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)"
	chmod +x "$(DB_INFRA_SCRIPTS_PATH)/postgres-entrypoint.sh" && \
	chmod +x "$(DB_INFRA_SCRIPTS_PATH)/run-migration.sh" && \
	docker compose --env-file ./.env --profile prod up -d

build-release:
	@printf "\n$(YELLOW)Tagging images to be released.$(NC)\n"
	docker tag api-api ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_api:$(BUILD_VERSION) && \
	docker tag api-caddy_prod ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_proxy:$(BUILD_VERSION)

	@printf "\n$(CYAN)Pushing release to GitHub registry.$(NC)\n"
	docker push ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_api:$(BUILD_VERSION) && \
	docker push ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_proxy:$(BUILD_VERSION)

build-prod-force:
	docker compose --env-file ./.env --profile prod up -d --force-recreate caddy_prod
