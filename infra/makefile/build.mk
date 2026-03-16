.PHONY: build-local build-ci build-prod build-release build-deploy build-local-restart build-prod-force build-fresh ensure-caddy-net ensure-base-images ensure-builder-base-image ensure-runtime-base-image build-base-images push-base-images

BUILD_VERSION ?= latest
BASE_IMAGE_VERSION ?= 1.26.1-alpine3.23-r1
BUILD_CADDY_NET := caddy_net
BUILD_PACKAGE_OWNER := oullin
BUILD_BASE_IMAGES_DIR := $(ROOT_PATH)/infra/docker/base-images
BUILD_BASE_BUILDER_IMAGE := oullin-api-builder-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_RUNTIME_IMAGE := oullin-api-runtime-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_BUILDER_IMAGE_REMOTE := ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin-api-builder-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_RUNTIME_IMAGE_REMOTE := ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin-api-runtime-base:$(BASE_IMAGE_VERSION)
DB_INFRA_ROOT_PATH ?= $(ROOT_PATH)/database/infra
DB_INFRA_SCRIPTS_PATH ?= $(DB_INFRA_ROOT_PATH)/scripts

ensure-caddy-net:
	docker network inspect caddy_net >/dev/null 2>&1 || docker network create caddy_net

ensure-base-images: ensure-builder-base-image ensure-runtime-base-image

ensure-builder-base-image:
	@docker image inspect "$(BUILD_BASE_BUILDER_IMAGE)" >/dev/null 2>&1 || \
		docker build -f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.builder" -t "$(BUILD_BASE_BUILDER_IMAGE)" "$(BUILD_BASE_IMAGES_DIR)"

ensure-runtime-base-image:
	@docker image inspect "$(BUILD_BASE_RUNTIME_IMAGE)" >/dev/null 2>&1 || \
		docker build -f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.runtime" -t "$(BUILD_BASE_RUNTIME_IMAGE)" "$(BUILD_BASE_IMAGES_DIR)"

build-base-images:
	@printf "\n$(CYAN)Building reproducible API base images$(NC)\n"
	docker build \
		-f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.builder" \
		-t "$(BUILD_BASE_BUILDER_IMAGE)" \
		"$(BUILD_BASE_IMAGES_DIR)"
	docker build \
		-f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.runtime" \
		-t "$(BUILD_BASE_RUNTIME_IMAGE)" \
		"$(BUILD_BASE_IMAGES_DIR)"

push-base-images:
	@$(MAKE) ensure-base-images
	@printf "\n$(CYAN)Tagging API base images for GitHub registry$(NC)\n"
	docker tag "$(BUILD_BASE_BUILDER_IMAGE)" "$(BUILD_BASE_BUILDER_IMAGE_REMOTE)"
	docker tag "$(BUILD_BASE_RUNTIME_IMAGE)" "$(BUILD_BASE_RUNTIME_IMAGE_REMOTE)"
	@printf "\n$(CYAN)Pushing API base images$(NC)\n"
	docker push "$(BUILD_BASE_BUILDER_IMAGE_REMOTE)"
	docker push "$(BUILD_BASE_RUNTIME_IMAGE_REMOTE)"
	@printf "\n$(CYAN)Published digests$(NC)\n"
	@docker image inspect "$(BUILD_BASE_BUILDER_IMAGE_REMOTE)" --format='builder={{index .RepoDigests 0}}'
	@docker image inspect "$(BUILD_BASE_RUNTIME_IMAGE_REMOTE)" --format='runtime={{index .RepoDigests 0}}'

build-fresh:
	$(MAKE) fresh && \
	$(MAKE) db:fresh && \
	$(MAKE) db:migrate && \
	$(MAKE) db:seed

build-local:
	$(MAKE) ensure-caddy-net
	$(MAKE) ensure-db-volume
	$(MAKE) ensure-base-images
	docker compose --profile local up --build -d

build-local-restart:
	$(MAKE) ensure-db-volume && \
	$(MAKE) ensure-base-images && \
	docker compose --profile local down && \
	docker compose --profile local up --build -d

build-ci:
	@printf "\n$(CYAN)Building production images for CI$(NC)\n"
	# This 'build' command only builds the images; it does not run them.
	# Build only services that have custom dockerfiles (not pre-built images)
	@$(MAKE) ensure-base-images
	@docker compose build api caddy_prod

# --- Deprecated
#     We should always deploy builds from the CI and not build again in servers.
build-prod:
	@$(MAKE) ensure-db-volume
	@$(MAKE) ensure-base-images
	@DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
	DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
	DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" \
	docker compose --profile prod up --build -d

build-deploy:
	@DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
	DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
	DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)"
	@$(MAKE) ensure-db-volume
	@$(MAKE) ensure-base-images
	chmod +x "$(DB_INFRA_SCRIPTS_PATH)/postgres-entrypoint.sh" && \
	chmod +x "$(DB_INFRA_SCRIPTS_PATH)/run-migration.sh"
	@echo "Starting database service..."
	docker compose --env-file ./.env up $(DB_DOCKER_SERVICE_NAME) -d
	@echo "Waiting for database to be healthy..."
	@attempt=0; max_attempts=30; \
	while [ $$attempt -lt $$max_attempts ]; do \
		if docker inspect --format="{{.State.Health.Status}}" $(DB_DOCKER_CONTAINER_NAME) 2>/dev/null | grep -q "healthy"; then \
			echo "Database is healthy"; \
			break; \
		fi; \
		attempt=$$((attempt + 1)); \
		if [ $$attempt -eq $$max_attempts ]; then \
			echo "Database failed to become healthy after 60 seconds" >&2; \
			exit 1; \
		fi; \
		sleep 2; \
	done
	@echo "Running migrations..."
	$(MAKE) db:migrate
	@echo "Starting remaining services..."
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
