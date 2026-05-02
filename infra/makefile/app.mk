# -------------------------------------------------------------------------------------------------------------------- #
# Application Management Targets
# -------------------------------------------------------------------------------------------------------------------- #

# -------------------------------------------------------------------------------------------------------------------- #
# Configuration Variables
# -------------------------------------------------------------------------------------------------------------------- #

ROOT_PATH           := $(shell pwd)
DB_SECRETS_DIR      := $(ROOT_PATH)/database/infra/secrets

# "auto" lets the local Go installation download the toolchain required by go.mod,
# so developers don't need to install the exact Go version manually.
# Override with a pinned version (e.g., GO_LOCAL_TOOLCHAIN=go1.26.1) for deterministic builds.
# Note: docker-compose reads GO_LOCAL_TOOLCHAIN from the environment separately.
GO_LOCAL_TOOLCHAIN  ?= auto
GOIMPORTS_VERSION   ?= v0.43.0
GOBIN               := $(shell go env GOPATH)/bin

DB_SECRET_USERNAME  ?= $(DB_SECRETS_DIR)/pg_username
DB_SECRET_PASSWORD  ?= $(DB_SECRETS_DIR)/pg_password
DB_SECRET_DBNAME    ?= $(DB_SECRETS_DIR)/pg_dbname

# -------------------------------------------------------------------------------------------------------------------- #
# PHONY Targets
# -------------------------------------------------------------------------------------------------------------------- #

.PHONY: fresh destroy audit watch format run-cli test-all run-cli-docker run-metal install-air install-goimports

run-cli run-cli-docker: export DB_SECRET_USERNAME := $(value DB_SECRET_USERNAME)
run-cli run-cli-docker: export DB_SECRET_PASSWORD := $(value DB_SECRET_PASSWORD)
run-cli run-cli-docker: export DB_SECRET_DBNAME := $(value DB_SECRET_DBNAME)

# -------------------------------------------------------------------------------------------------------------------- #
# Code Quality Commands
# -------------------------------------------------------------------------------------------------------------------- #

format:
	@if [ ! -f $(GOBIN)/goimports ]; then \
		printf "\n  $(YELLOW)goimports not found — installing now...$(NC)\n"; \
		$(MAKE) install-goimports; \
	fi
	@printf "\n  ..... $(CYAN)gofmt & goimports commands have started.$(NC)\n"
	@git ls-files -z '*.go' | xargs -0 env GOTOOLCHAIN=$(GO_LOCAL_TOOLCHAIN) gofmt -w -s
	@git ls-files -z '*.go' | xargs -0 $(GOBIN)/goimports -w -local github.com/oullin
	@printf "\n  ..... $(GREEN)Formatting finished.$(NC)\n\n"

audit:
	$(call external_deps,'.')
	$(call external_deps,'./app/...')
	$(call external_deps,'./database/...')
	$(call external_deps,'./docs/...')

test-all:
	@GOTOOLCHAIN=$(GO_LOCAL_TOOLCHAIN) go test ./...

# -------------------------------------------------------------------------------------------------------------------- #
# Docker Management Commands
# -------------------------------------------------------------------------------------------------------------------- #

fresh:
	docker compose down --volumes --rmi all --remove-orphans
	docker ps

destroy:
	docker compose down --remove-orphans && \
	docker container prune -f && \
	docker image prune -f && \
	docker volume prune -f && \
	docker network prune -f && \
	docker system prune -a --volumes -f && \
	docker ps -aq | xargs --no-run-if-empty docker stop && \
	docker ps -aq | xargs --no-run-if-empty docker rm && \
	docker ps

# -------------------------------------------------------------------------------------------------------------------- #
# Development Tools
# -------------------------------------------------------------------------------------------------------------------- #

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(APP_PATH) && GOTOOLCHAIN=$(GO_LOCAL_TOOLCHAIN) air -d

install-air:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	@echo "Installing air ..."
	@GOTOOLCHAIN=$(GO_LOCAL_TOOLCHAIN) go install github.com/air-verse/air@latest

install-goimports:
	@echo "Installing goimports $(GOIMPORTS_VERSION) ..."
	@GOTOOLCHAIN=$(GO_LOCAL_TOOLCHAIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)

# -------------------------------------------------------------------------------------------------------------------- #
# CLI Commands
# -------------------------------------------------------------------------------------------------------------------- #

run-cli:
	@missing_values=""; \
	missing_files=""; \
	check_secret() { \
		secret_name="$$1"; \
		secret_value="$$2"; \
		if [ -z "$$secret_value" ]; then \
			if [ -z "$$missing_values" ]; then \
				missing_values="  - $$secret_name"; \
			else \
				missing_values="$$missing_values\n  - $$secret_name"; \
			fi; \
		else \
			case "$$secret_value" in \
				/*|./*|../*) \
					if [ ! -f "$$secret_value" ]; then \
						if [ -z "$$missing_files" ]; then \
							missing_files="  - $$secret_name ($$secret_value)"; \
						else \
							missing_files="$$missing_files\n  - $$secret_name ($$secret_value)"; \
						fi; \
					fi; \
					;; \
				*) \
					if [ -z "$$missing_files" ]; then \
						missing_files="  - $$secret_name (literal value — must be a file path)"; \
					else \
						missing_files="$$missing_files\n  - $$secret_name (literal value — must be a file path)"; \
					fi; \
					;; \
			esac; \
		fi; \
	}; \
	check_secret DB_SECRET_USERNAME "$$DB_SECRET_USERNAME"; \
	check_secret DB_SECRET_PASSWORD "$$DB_SECRET_PASSWORD"; \
	check_secret DB_SECRET_DBNAME "$$DB_SECRET_DBNAME"; \
	if [ -n "$$missing_values" ]; then \
		printf "\n$(RED)❌ Missing secret values:$(NC)\n"; \
		printf '%b\n' "$$missing_values"; \
		printf "  Provide them via environment variables or override them when invoking $(BOLD)make run-cli$(NC).\n\n"; \
		exit 1; \
	fi; \
	if [ -n "$$missing_files" ]; then \
		printf "\n$(RED)❌ Secret file paths not found:$(NC)\n"; \
		printf '%b\n' "$$missing_files"; \
		printf "  Ensure the files exist or adjust the overrides before running $(BOLD)make run-cli$(NC).\n\n"; \
		exit 1; \
	fi
	@printf "\n$(GREEN)🔒 Running CLI with secrets from:$(NC)\n"
	@DB_SECRET_USERNAME_DISPLAY=`case "$$DB_SECRET_USERNAME" in \
		/*|./*|../*) printf '%s' "$$DB_SECRET_USERNAME";; \
		"") printf '<unset>';; \
		*) printf '<redacted>';; \
		esac`; \
	printf "           DB_SECRET_USERNAME=%s\n" "$$DB_SECRET_USERNAME_DISPLAY"
	@DB_SECRET_PASSWORD_DISPLAY=`case "$$DB_SECRET_PASSWORD" in \
		/*|./*|../*) printf '%s' "$$DB_SECRET_PASSWORD";; \
		"") printf '<unset>';; \
		*) printf '<redacted>';; \
		esac`; \
	printf "           DB_SECRET_PASSWORD=%s\n" "$$DB_SECRET_PASSWORD_DISPLAY"
	@DB_SECRET_DBNAME_DISPLAY=`case "$$DB_SECRET_DBNAME" in \
		/*|./*|../*) printf '%s' "$$DB_SECRET_DBNAME";; \
		"") printf '<unset>';; \
		*) printf '<redacted>';; \
		esac`; \
	printf "           DB_SECRET_DBNAME=%s\n\n" "$$DB_SECRET_DBNAME_DISPLAY"
	@status=0; \
	compose_cmd=""; \
	if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		compose_cmd="docker compose"; \
		printf "Using docker compose to run the CLI.\n"; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		compose_cmd="docker-compose"; \
		printf "Using docker-compose to run the CLI.\n"; \
	else \
		printf "\n$(RED)❌ Neither 'docker compose' nor 'docker-compose' is available.$(NC)\n"; \
		printf "   Install Docker Compose or run the CLI locally without containers.\n\n"; \
		exit 1; \
	fi; \
	$(DB_DOCKER_STATE_FUNCS) \
	$(MAKE) --no-print-directory ensure-base-images || status=$$?; \
	if [ $$status -eq 0 ] && [ "$$(db_running)" = "true" ] && [ "$$(db_health)" = "healthy" ]; then \
		printf "Database container $(DB_DOCKER_CONTAINER_NAME) is already healthy.\n"; \
	elif [ $$status -eq 0 ]; then \
		printf "Database container $(DB_DOCKER_CONTAINER_NAME) is not ready. Starting $(DB_DOCKER_SERVICE_NAME)...\n"; \
		$(MAKE) --no-print-directory ensure-db-volume || status=$$?; \
		if [ $$status -eq 0 ]; then \
			$$compose_cmd up -d $(DB_DOCKER_SERVICE_NAME) || status=$$?; \
		fi; \
		if [ $$status -eq 0 ]; then \
			printf "Waiting for database to become healthy...\n"; \
			attempt=0; max_attempts=30; \
			while [ $$attempt -lt $$max_attempts ]; do \
				if [ "$$(db_running)" = "true" ] && [ "$$(db_health)" = "healthy" ]; then \
					printf "Database is healthy.\n"; \
					break; \
				fi; \
				attempt=$$((attempt + 1)); \
				if [ $$attempt -eq $$max_attempts ]; then \
					printf "\n$(RED)❌ Database failed to become healthy after 60 seconds.$(NC)\n"; \
					status=1; \
					break; \
				fi; \
				sleep 2; \
			done; \
		fi; \
	fi; \
	if [ $$status -eq 0 ]; then \
		$(MAKE) --no-print-directory build-cli-docker || status=$$?; \
	fi; \
	if [ $$status -eq 0 ]; then \
		$$compose_cmd run --rm --no-deps api-runner $(CLI_DOCKER_BINARY_CONTAINER) || status=$$?; \
	fi; \
	if [ $$status -ne 0 ]; then \
		printf "\n$(RED)❌ CLI exited with status $$status.$(NC)\n"; \
		exit $$status; \
	fi

run-cli-docker: ensure-base-images
	$(MAKE) run-cli

run-metal:
	@GOTOOLCHAIN=$(GO_LOCAL_TOOLCHAIN) go run metal/cli/main.go
