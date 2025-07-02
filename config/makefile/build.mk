.PHONY: build\:local build\:prod build\:release

BUILD_VERSION := 0.0.2

build\:local:
	docker compose --profile local up --build -d

build\:prod:
	@printf "\n$(CYAN)docker compose --profile prod up --build -d$(NC)\n"
	# --- The following lines take the variables passed to 'make' and export them
	#     into the shell environment for only the docker-compose command.
	#     These variable names now EXACTLY match what the Go application expects.
	@POSTGRES_USER_SECRET_PATH="$(POSTGRES_USER_SECRET_PATH)" \
	POSTGRES_PASSWORD_SECRET_PATH="$(POSTGRES_PASSWORD_SECRET_PATH)" \
	POSTGRES_DB_SECRET_PATH="$(POSTGRES_DB_SECRET_PATH)" \
	ENV_DB_USER_NAME="$(ENV_DB_USER_NAME)" \
	ENV_DB_USER_PASSWORD="$(ENV_DB_USER_PASSWORD)" \
	ENV_DB_DATABASE_NAME="$(ENV_DB_DATABASE_NAME)" \
	docker compose --profile prod up --build -d

build\:release:
	@printf "\n$(YELLOW)Tagging images to be released.$(NC)\n"
	docker tag api-api ghcr.io/gocanto/oullin_api:$(BUILD_VERSION) && \
	docker tag api-caddy_prod ghcr.io/gocanto/oullin_proxy:$(BUILD_VERSION)

	@printf "\n$(CYAN)Pushing release to GitHub registry.$(NC)\n"
	docker push ghcr.io/gocanto/oullin_api:$(BUILD_VERSION) && \
	docker push ghcr.io/gocanto/oullin_proxy:$(BUILD_VERSION)
