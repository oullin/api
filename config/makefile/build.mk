.PHONY: build\:local build\:prod build\:release

BUILD_VERSION ?= latest
BUILD_PACKAGE_OWNER := oullin

build\:local:
	docker compose --profile local up --build -d

build\:prod:
	@printf "\n$(CYAN)docker compose --profile prod up --build -d$(NC)\n"
	# --- The following lines take the variables passed to 'make' and export them
	#     into the shell environment for only the docker-compose command.
	#     These variable names now EXACTLY match what the Go application expects.
	@DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
	DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
	DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" \
	docker compose --profile prod up --build -d

build\:release:
	@printf "\n$(YELLOW)Tagging images to be released.$(NC)\n"
	docker tag api-api ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_api:$(BUILD_VERSION) && \
	docker tag api-caddy_prod ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_proxy:$(BUILD_VERSION)

	@printf "\n$(CYAN)Pushing release to GitHub registry.$(NC)\n"
	docker push ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_api:$(BUILD_VERSION) && \
	docker push ghcr.io/$(BUILD_PACKAGE_OWNER)/oullin_proxy:$(BUILD_VERSION)
