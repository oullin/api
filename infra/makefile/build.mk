.PHONY: build-local watch-local build-ci build-prod build-release build-deploy build-local-restart build-prod-force build-fresh ensure-caddy-net ensure-base-images ensure-builder-base-image ensure-runtime-base-image build-base-images push-base-images generate-apk-checksums build-cli-docker prewarm-cli-docker

BUILD_VERSION ?= latest
BASE_GO_VERSION ?= 1.26.1
BASE_ALPINE_VERSION ?= 3.23
BASE_IMAGE_REVISION ?= 2
BASE_GO_IMAGE_VARIANT ?= alpine$(BASE_ALPINE_VERSION)
BASE_GO_IMAGE_DIGEST ?= sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039
BASE_ALPINE_IMAGE_DIGEST ?= sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
BASE_APK_BASE_URL ?= https://dl-cdn.alpinelinux.org/alpine/v$(BASE_ALPINE_VERSION)/main
BASE_IMAGE_VERSION ?= $(BASE_GO_VERSION)-alpine$(BASE_ALPINE_VERSION)-r$(BASE_IMAGE_REVISION)
BUILD_CADDY_NET := caddy_net
BUILD_PACKAGE_OWNER := oullin
BUILD_BASE_IMAGES_DIR := $(ROOT_PATH)/infra/docker/base-images
BUILD_BASE_BUILDER_IMAGE := oullin-api-builder-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_RUNTIME_IMAGE := oullin-api-runtime-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_BUILDER_IMAGE_REMOTE := ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin-api-builder-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_RUNTIME_IMAGE_REMOTE := ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin-api-runtime-base:$(BASE_IMAGE_VERSION)
BUILD_BASE_BUILDER_ARGS := --build-arg GO_VERSION=$(BASE_GO_VERSION) --build-arg GO_IMAGE_VARIANT=$(BASE_GO_IMAGE_VARIANT) --build-arg GO_IMAGE_DIGEST=$(BASE_GO_IMAGE_DIGEST) --build-arg APK_BASE_URL=$(BASE_APK_BASE_URL)
BUILD_BASE_RUNTIME_ARGS := --build-arg ALPINE_VERSION=$(BASE_ALPINE_VERSION) --build-arg ALPINE_IMAGE_DIGEST=$(BASE_ALPINE_IMAGE_DIGEST) --build-arg APK_BASE_URL=$(BASE_APK_BASE_URL)
DB_INFRA_ROOT_PATH ?= $(ROOT_PATH)/database/infra
DB_INFRA_SCRIPTS_PATH ?= $(DB_INFRA_ROOT_PATH)/scripts
CLI_DOCKER_BINARY_HOST := $(ROOT_PATH)/bin/metal-cli
CLI_DOCKER_BINARY_CONTAINER := /app/bin/metal-cli
CLI_DOCKER_BUILD_INPUTS := $(shell git ls-files '*.go' go.mod go.sum 2>/dev/null)

build-local build-local-restart build-ci build-prod build-deploy run-cli run-cli-docker build-cli-docker prewarm-cli-docker: export BASE_IMAGE_VERSION := $(BASE_IMAGE_VERSION)
build-prod build-deploy: export DB_SECRET_USERNAME := $(value DB_SECRET_USERNAME)
build-prod build-deploy: export DB_SECRET_PASSWORD := $(value DB_SECRET_PASSWORD)
build-prod build-deploy: export DB_SECRET_DBNAME := $(value DB_SECRET_DBNAME)

ensure-caddy-net:
	docker network inspect caddy_net >/dev/null 2>&1 || docker network create caddy_net

ensure-base-images: ensure-builder-base-image ensure-runtime-base-image

ensure-builder-base-image:
	@docker image inspect "$(BUILD_BASE_BUILDER_IMAGE)" >/dev/null 2>&1 || \
		docker build $(BUILD_BASE_BUILDER_ARGS) -f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.builder" -t "$(BUILD_BASE_BUILDER_IMAGE)" "$(BUILD_BASE_IMAGES_DIR)"

ensure-runtime-base-image:
	@docker image inspect "$(BUILD_BASE_RUNTIME_IMAGE)" >/dev/null 2>&1 || \
		docker build $(BUILD_BASE_RUNTIME_ARGS) -f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.runtime" -t "$(BUILD_BASE_RUNTIME_IMAGE)" "$(BUILD_BASE_IMAGES_DIR)"

generate-apk-checksums:
	@printf "\n$(CYAN)Generating APK checksum files$(NC)\n"
	APK_BASE_URL=$(BASE_APK_BASE_URL) "$(BUILD_BASE_IMAGES_DIR)/generate-checksums.sh"

build-base-images:
	@printf "\n$(CYAN)Building reproducible API base images$(NC)\n"
	docker build \
		$(BUILD_BASE_BUILDER_ARGS) \
		-f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.builder" \
		-t "$(BUILD_BASE_BUILDER_IMAGE)" \
		"$(BUILD_BASE_IMAGES_DIR)"
	docker build \
		$(BUILD_BASE_RUNTIME_ARGS) \
		-f "$(BUILD_BASE_IMAGES_DIR)/Dockerfile.runtime" \
		-t "$(BUILD_BASE_RUNTIME_IMAGE)" \
		"$(BUILD_BASE_IMAGES_DIR)"

build-cli-docker: $(CLI_DOCKER_BINARY_HOST)
	@printf "  $(CYAN)Docker CLI binary ready at %s.$(NC)\n" "$(CLI_DOCKER_BINARY_HOST)"

prewarm-cli-docker:
	@printf "\n$(CYAN)Warming Docker CLI caches and binary$(NC)\n"
	@printf "  This does not start or wait for the database.\n"
	@$(MAKE) --no-print-directory build-cli-docker

$(CLI_DOCKER_BINARY_HOST): $(CLI_DOCKER_BUILD_INPUTS) | ensure-base-images
	@mkdir -p "$(dir $@)"
	@status=0; \
	if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		printf "Building Docker CLI binary at $(CLI_DOCKER_BINARY_CONTAINER).\n"; \
		docker compose run --rm --no-deps api-runner sh -lc 'go build -o "$(CLI_DOCKER_BINARY_CONTAINER)" ./metal/cli/main.go' || status=$$?; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		printf "Building Docker CLI binary at $(CLI_DOCKER_BINARY_CONTAINER).\n"; \
		docker-compose run --rm --no-deps api-runner sh -lc 'go build -o "$(CLI_DOCKER_BINARY_CONTAINER)" ./metal/cli/main.go' || status=$$?; \
	else \
		printf "\n$(RED)❌ Neither 'docker compose' nor 'docker-compose' is available.$(NC)\n"; \
		printf "   Install Docker Compose or run the CLI locally without containers.\n\n"; \
		exit 1; \
	fi; \
	if [ $$status -ne 0 ]; then \
		printf "\n$(RED)❌ Failed to build the Docker CLI binary (status $$status).$(NC)\n"; \
		exit $$status; \
	fi

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

watch-local:
	$(MAKE) ensure-caddy-net
	$(MAKE) ensure-db-volume
	$(MAKE) ensure-base-images
	docker compose --profile local up

build-local-restart:
	$(MAKE) ensure-db-volume
	$(MAKE) ensure-base-images
	docker compose --profile local down
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
	@docker compose --profile prod up --build -d

build-deploy:
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
