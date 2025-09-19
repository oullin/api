.PHONY: fresh destroy audit watch format run-cli test-all run-cli-docker run-cli-local

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
	@if [ -z "$(DB_SECRET_USERNAME)" ] || [ -z "$(DB_SECRET_PASSWORD)" ] || [ -z "$(DB_SECRET_DBNAME)" ]; then \
    	  printf "\n$(RED)⚠️ Usage: make run-cli \n$(NC)"; \
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
    	docker compose run --rm api-runner go run ./metal/cli/main.go

run-cli-docker:
	make run-cli DB_SECRET_USERNAME=./database/infra/secrets/pg_username DB_SECRET_PASSWORD=./database/infra/secrets/pg_password DB_SECRET_DBNAME=./database/infra/secrets/pg_dbname

test-all:
	go test ./...

run-cli-local:
	go run metal/cli/main.go
