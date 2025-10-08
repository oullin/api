.PHONY: db\:sh db\:up db\:down db\:logs db\:bash db\:fresh
.PHONY: db\:secure db\:seed db\:import db\:migrate db\:migrate\:create db\:migrate\:force db\:rollback db\:chmod

# --- Docker Services
DB_API_RUNNER_SERVICE := api-runner
DB_DOCKER_SERVICE_NAME := api-db
DB_DOCKER_CONTAINER_NAME := oullin_db
DB_MIGRATE_SERVICE_NAME := api-db-migrate

# --- Paths
#     Define root paths for clarity. Assumes ROOT_PATH is exported or defined.
DB_INFRA_ROOT_PATH ?= $(ROOT_PATH)/database/infra
DB_INFRA_SSL_PATH := $(DB_INFRA_ROOT_PATH)/ssl
DB_INFRA_SCRIPTS_PATH ?= $(DB_INFRA_ROOT_PATH)/scripts

# --- Migrations
DB_MIGRATE_DOCKER_ENV_FLAGS = -e ENV_DB_HOST=$(DB_DOCKER_SERVICE_NAME) \
                              -e ENV_DB_SSL_MODE=require

# --- SSL Certificate Files
DB_INFRA_SERVER_CRT := $(DB_INFRA_SSL_PATH)/server.crt
DB_INFRA_SERVER_CSR := $(DB_INFRA_SSL_PATH)/server.csr
DB_INFRA_SERVER_KEY := $(DB_INFRA_SSL_PATH)/server.key

db\:sh:
	chmod +x $(DB_INFRA_SCRIPTS_PATH)/healthcheck.sh
	chmod +x $(DB_INFRA_SCRIPTS_PATH)/run-migration.sh

db\:up:
	docker compose up $(DB_DOCKER_SERVICE_NAME) -d

db\:down:
	docker compose stop $(DB_DOCKER_SERVICE_NAME)

db\:logs:
	docker logs -f $(DB_DOCKER_CONTAINER_NAME)

db\:bash:
	docker exec -it $(DB_DOCKER_CONTAINER_NAME) bash

db\:fresh:
	make db:delete
	make db:up

db\:delete:
	docker compose down -v --remove-orphans

db\:chmod:
	sudo chmod 600 $(DB_INFRA_SERVER_KEY)
	sudo chmod 644 $(DB_INFRA_SERVER_CRT)

db\:secure:
	rm -f $(DB_INFRA_SERVER_CRT) $(DB_INFRA_SERVER_CSR) $(DB_INFRA_SERVER_KEY)
	openssl genpkey -algorithm RSA -out $(DB_INFRA_SERVER_KEY)
	openssl req -new -key $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CSR) -subj "/CN=oullin-db-ssl"
	openssl x509 -req -days 365 -in $(DB_INFRA_SERVER_CSR) -signkey $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CRT)

db\:seed:
	docker compose --env-file ./.env run --rm $(DB_MIGRATE_DOCKER_ENV_FLAGS) $(DB_API_RUNNER_SERVICE) \
		go run ./database/seeder/main.go

db\:import:
	@if [ -z "$(file)" ]; then \
	echo "usage: make db:import file=path/to/seed.sql"; \
	exit 1; \
	fi
	docker compose --env-file ./.env run --rm $(DB_MIGRATE_DOCKER_ENV_FLAGS) $(DB_API_RUNNER_SERVICE) \
		go run ./database/seeder/importer/main --file $(file)

# -------------------------------------------------------------------------------------------------------------------- #
# --- Migrations
# -------------------------------------------------------------------------------------------------------------------- #
db\:migrate:
	docker compose run --rm $(DB_MIGRATE_SERVICE_NAME) up

db\:rollback:
	docker compose run --rm $(DB_MIGRATE_SERVICE_NAME) down 1

db\:migrate\:create:
	docker compose run --rm $(DB_MIGRATE_SERVICE_NAME) create -ext sql -dir /migrations -seq $(name)

db\:migrate\:force:
	docker compose run --rm $(DB_MIGRATE_SERVICE_NAME) force $(version)
