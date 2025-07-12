.PHONY: fresh audit watch format

format:
	gofmt -w -s .

fresh:
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
	cd $(APP_PATH) && air

install-air:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	@echo "Installing air ..."
	@go install github.com/air-verse/air@latest

YEL := \033[0;33m
GRN := \033[0;32m
NC  := \033[0m

app-cli:
	@if [ -z "$(DB_SECRET_USERNAME)" ] || [ -z "$(DB_SECRET_PASSWORD)" ] || [ -z "$(DB_SECRET_DBNAME)" ]; then \
    	  printf "\n$(RED)⚠️ Usage: make app-cli \n$(NC)"; \
    	  printf "         DB_SECRET_USERNAME=/path/to/pg_username\n"; \
    	  printf "         DB_SECRET_PASSWORD=/path/to/pg_password\n"; \
    	  printf "         DB_SECRET_DBNAME=/path/to/pg_dbname\n\n"; \
    	  printf "\n------------------------------------------------------\n\n"; \
    	  exit 1; \
    	fi; \
    	printf "\n$(GREEN)🔒 Running CLI with secrets from:$(NC)\n"; \
    	printf "           DB_SECRET_USERNAME=$(DB_SECRET_USERNAME)\n"; \
    	printf "           DB_SECRET_PASSWORD=$(DB_SECRET_PASSWORD)\n"; \
    	printf "           DB_SECRET_DBNAME=$(DB_SECRET_DBNAME)\n\n"; \
    	DB_SECRET_USERNAME="$(DB_SECRET_USERNAME)" \
    	DB_SECRET_PASSWORD="$(DB_SECRET_PASSWORD)" \
    	DB_SECRET_DBNAME="$(DB_SECRET_DBNAME)" \
    	docker compose run --rm api-runner go run $(ROOT_PATH)/cli/main.go
