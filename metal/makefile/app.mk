.PHONY: fresh destroy audit watch format run-cli test-all run-cli-docker run-metal open-prometheus

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
	@DB_SECRET_USERNAME_DISPLAY=`case "$(DB_SECRET_USERNAME)" in \
		/*|./*|../*) printf '%s' "$(DB_SECRET_USERNAME)";; \
		"") printf '<unset>';; \
		*) printf '<redacted>';; \
		esac`; \
	printf "           DB_SECRET_USERNAME=%s\n" "$$DB_SECRET_USERNAME_DISPLAY"
	@DB_SECRET_PASSWORD_DISPLAY=`case "$(DB_SECRET_PASSWORD)" in \
		/*|./*|../*) printf '%s' "$(DB_SECRET_PASSWORD)";; \
		"") printf '<unset>';; \
		*) printf '<redacted>';; \
		esac`; \
	printf "           DB_SECRET_PASSWORD=%s\n" "$$DB_SECRET_PASSWORD_DISPLAY"
	@DB_SECRET_DBNAME_DISPLAY=`case "$(DB_SECRET_DBNAME)" in \
		/*|./*|../*) printf '%s' "$(DB_SECRET_DBNAME)";; \
		"") printf '<unset>';; \
		*) printf '<redacted>';; \
		esac`; \
	printf "           DB_SECRET_DBNAME=%s\n\n" "$$DB_SECRET_DBNAME_DISPLAY"
	@status=0; \
	if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		printf "Using docker compose to run the CLI.\n"; \
		DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" docker compose run --rm api-runner go run ./metal/cli/main.go || status=$$?; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		printf "Using docker-compose to run the CLI.\n"; \
		DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" docker-compose run --rm api-runner go run ./metal/cli/main.go || status=$$?; \
	else \
		printf "\n$(RED)❌ Neither 'docker compose' nor 'docker-compose' is available.$(NC)\n"; \
		printf "   Install Docker Compose or run the CLI locally without containers.\n\n"; \
		exit 1; \
	fi; \
	if [ $$status -ne 0 ]; then \
		printf "\n$(RED)❌ CLI exited with status $$status.$(NC)\n"; \
		exit $$status; \
	fi
run-cli-docker:
	make run-cli DB_SECRET_USERNAME=$(DB_SECRET_USERNAME) DB_SECRET_PASSWORD=$(DB_SECRET_PASSWORD) DB_SECRET_DBNAME=$(DB_SECRET_DBNAME)

test-all:
	go test ./...

run-metal:
	go run metal/cli/main.go

open-prometheus:
	@url="http://localhost:9090"; \
	printf "Attempting to open Prometheus dashboard at %s\\n" "$$url"; \
	if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open "$$url"; \
	elif command -v sensible-browser >/dev/null 2>&1; then \
		sensible-browser "$$url"; \
	elif command -v w3m >/dev/null 2>&1; then \
		w3m "$$url"; \
	elif command -v lynx >/dev/null 2>&1; then \
		lynx "$$url"; \
	elif command -v links >/dev/null 2>&1; then \
		links "$$url"; \
	elif command -v python3 >/dev/null 2>&1; then \
		python3 -m webbrowser "$$url"; \
	else \
		printf "Unable to locate a browser command. Please open %s manually.\\n" "$$url"; \
	fi
