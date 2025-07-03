.PHONY: db\:sh db\:up db\:down db\:logs db\:bash db\:fresh
.PHONY: db\:secure db\:seed db\:migrate db\:migrate\:create db\:migrate\:force db\:rollback

# --- Docker Services
DB_API_RUNNER_SERVICE := api-runner
DB_DOCKER_SERVICE_NAME := api-db
DB_DOCKER_CONTAINER_NAME := oullin_db
DB_MIGRATE_SERVICE_NAME := api-db-migrate

# --- Paths
#     Define root paths for clarity. Assumes ROOT_PATH is exported or defined.
DB_INFRA_ROOT_PATH := $(ROOT_PATH)/database/infra
DB_INFRA_SSL_PATH := $(DB_INFRA_ROOT_PATH)/ssl
DB_INFRA_SCRIPTS_PATH := $(DB_INFRA_ROOT_PATH)/scripts
# --- Secrets
ENV_DB_INFRA_SECRETS_PATH ?= $(DB_INFRA_ROOT_PATH)/secrets
DB_SECRET_FILE_USERNAME := $(ENV_DB_INFRA_SECRETS_PATH)/postgres_user
DB_SECRET_FILE_PASSWORD := $(ENV_DB_INFRA_SECRETS_PATH)/postgres_password
DB_SECRET_FILE_DBNAME   := $(ENV_DB_INFRA_SECRETS_PATH)/postgres_db

DB_SECRET_FILE_BLOCK ?= -e ENV_DB_HOST=$(DB_DOCKER_SERVICE_NAME) \
						-e POSTGRES_USER_SECRET_PATH=$(DB_SECRET_FILE_USERNAME) \
                        -e POSTGRES_PASSWORD_SECRET_PATH=$(DB_SECRET_FILE_PASSWORD) \
                        -e POSTGRES_DB_SECRET_PATH=$(DB_SECRET_FILE_DBNAME)

DB_MIGRATE_URL=postgres://$(DB_DOCKER_SERVICE_NAME):5432/$(shell cat $(DB_SECRET_FILE_DBNAME))?sslmode=require

DB_MIGRATE_DOCKER_ENV_FLAGS = -e ENV_DB_HOST=$(DB_DOCKER_SERVICE_NAME) \
                              -e ENV_DB_SSL_MODE=require \
                              -e DATABASE_URL=$(DB_MIGRATE_URL)

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

db\:secure:
	rm -f $(DB_INFRA_SERVER_CRT) $(DB_INFRA_SERVER_CSR) $(DB_INFRA_SERVER_KEY)
	openssl genpkey -algorithm RSA -out $(DB_INFRA_SERVER_KEY)
	openssl req -new -key $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CSR) -subj "/CN=oullin-db-ssl"
	openssl x509 -req -days 365 -in $(DB_INFRA_SERVER_CSR) -signkey $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CRT)

db\:seed:
	docker compose run --rm $(DB_MIGRATE_DOCKER_ENV_FLAGS) $(DB_API_RUNNER_SERVICE) go run ./database/seeder/main.go

# -------------------------------------------------------------------------------------------------------------------- #
# --- Migrations
# -------------------------------------------------------------------------------------------------------------------- #
db\:migrate:
	docker compose run --rm $(DB_MIGRATE_DOCKER_ENV_FLAGS) $(DB_MIGRATE_SERVICE_NAME) -path /migrations -database "$$DATABASE_URL" up

db\:rollback:
	docker compose run --rm $(DB_MIGRATE_DOCKER_ENV_FLAGS) $(DB_MIGRATE_SERVICE_NAME) -path /migrations -database "$$DATABASE_URL" down 1

db\:migrate\:create:
	docker compose run --rm $(DB_MIGRATE_SERVICE_NAME) create -ext sql -dir /migrations -seq $(name)

db\:migrate\:force:
	docker compose run --rm $(DB_MIGRATE_SERVICE_NAME) force $(version)
