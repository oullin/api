.PHONY: fresh destroy audit watch format run-cli test-all run-cli-docker run-metal

DB_SECRET_USERNAME ?= ./database/infra/secrets/pg_username
DB_SECRET_PASSWORD ?= ./database/infra/secrets/pg_password
DB_SECRET_DBNAME   ?= ./database/infra/secrets/pg_dbname

format:
	gofmt -w -s .

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

audit:
	$(call external_deps,'.')
	$(call external_deps,'./app/...')
	$(call external_deps,'./database/...')
	$(call external_deps,'./docs/...')

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(APP_PATH) && air -d

install-air:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	@echo "Installing air ..."
	@go install github.com/air-verse/air@latest

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
			esac; \
		fi; \
	}; \
	check_secret DB_SECRET_USERNAME "$(DB_SECRET_USERNAME)"; \
	check_secret DB_SECRET_PASSWORD "$(DB_SECRET_PASSWORD)"; \
	check_secret DB_SECRET_DBNAME "$(DB_SECRET_DBNAME)"; \
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
	@printf "           DB_SECRET_USERNAME=$(DB_SECRET_USERNAME)\n"
	@printf "           DB_SECRET_PASSWORD=$(DB_SECRET_PASSWORD)\n"
	@printf "           DB_SECRET_DBNAME=$(DB_SECRET_DBNAME)\n\n"
	@if ! command -v docker >/dev/null 2>&1; then \
		printf "$(YELLOW)⚠️ Docker not available. Running CLI locally without Docker.\n$(NC)"; \
		DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
		DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
		DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" \
		go run ./metal/cli/main.go || { \
			status=$$?; \
			printf "\n$(RED)❌ CLI exited with status $$status.$(NC)\n"; \
			exit $$status; \
		}; \
	else \
		DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
		DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
		DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" \
		docker compose run --rm api-runner go run ./metal/cli/main.go || { \
			status=$$?; \
			printf "\n$(RED)❌ CLI exited with status $$status.$(NC)\n"; \
			exit $$status; \
		}; \
	fi
run-cli-docker:
	make run-cli DB_SECRET_USERNAME=$(DB_SECRET_USERNAME) DB_SECRET_PASSWORD=$(DB_SECRET_PASSWORD) DB_SECRET_DBNAME=$(DB_SECRET_DBNAME)

test-all:
	go test ./...

run-metal:
	go run metal/cli/main.go
