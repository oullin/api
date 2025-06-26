# --- Metadata
.PHONY: db\:local db\:up db\:ping db\:bash db\:fresh db\:logs
.PHONY: db\:delete db\:secure db\:secure\:show db\:chmod db\:seed
.PHONY: db\:migrate db\:rollback db\:migrate\:create db\:migrate\:force

# --- Docker
DB_DOCKER_SERVICE_NAME := "api-db"
DB_DOCKER_CONTAINER_NAME := oullin_db

# --- Paths
DB_SEEDER_ROOT_PATH := $(ROOT_PATH)/database/seeder
DB_INFRA_ROOT_PATH := $(ROOT_PATH)/database/infra
DB_INFRA_SSL_PATH := $(DB_INFRA_ROOT_PATH)/ssl
DB_INFRA_DATA_PATH := $(DB_INFRA_ROOT_PATH)/data

# --- SSL
DB_INFRA_SERVER_CRT := $(DB_INFRA_SSL_PATH)/server.crt
DB_INFRA_SERVER_CSR := $(DB_INFRA_SSL_PATH)/server.csr
DB_INFRA_SERVER_KEY := $(DB_INFRA_SSL_PATH)/server.key

# --- Migrations
DB_MIGRATE_PATH := $(DB_INFRA_ROOT_PATH)/migrations
DB_MIGRATE_VOL_MAP := $(DB_MIGRATE_PATH):$(DB_MIGRATE_PATH)

db\:local:
	# --- Works with your local PG installation.
	cd  $(EN_DB_BIN_DIR) && \
	./psql -h $(ENV_DB_HOST) -U $(ENV_DB_USER_NAME) -d $(ENV_DB_DATABASE_NAME) -p $(ENV_DB_PORT)

db\:seed:
	go run $(DB_SEEDER_ROOT_PATH)/main.go

db\:up:
	docker compose up $(DB_DOCKER_SERVICE_NAME) -d && \
	make db:logs

db\:ping:
	docker port $(DB_DOCKER_CONTAINER_NAME)

db\:bash:
	docker exec -it $(DB_DOCKER_CONTAINER_NAME) bash

db\:fresh:
	make db:delete && make db:up

db\:logs:
	docker logs -f $(DB_DOCKER_CONTAINER_NAME)

db\:delete:
	docker compose down $(DB_DOCKER_SERVICE_NAME) --remove-orphans && \
	sudo rm -rf $(DB_INFRA_DATA_PATH) && \
	docker ps

db\:secure:
	rm -rf $(DB_INFRA_SERVER_CRT) && rm -rf $(DB_INFRA_SERVER_CSR) && rm -rf $(DB_INFRA_SERVER_KEY) && \
	openssl genpkey -algorithm RSA -out $(DB_INFRA_SERVER_KEY) && \
    openssl req -new -key $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CSR) && \
    openssl x509 -req -days 365 -in $(DB_INFRA_SERVER_CSR) -signkey $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CRT) && \
    make db:chmod

db\:chmod:
	sudo chmod 600 $(DB_INFRA_SERVER_KEY) && sudo chmod 600 $(DB_INFRA_SERVER_CRT)

db\:secure\:show:
	docker exec -it $(DB_DOCKER_CONTAINER_NAME) ls -l /etc/ssl/private/server.key && \
	docker exec -it $(DB_DOCKER_CONTAINER_NAME) ls -l /etc/ssl/certs/server.crt

db\:migrate:
	@printf "\n$(BLUE)[DB]$(NC) Migration has started.\n"
	@docker run -v $(DB_MIGRATE_VOL_MAP) --network $(ROOT_NETWORK) migrate/migrate -verbose -path=$(DB_MIGRATE_PATH) -database $(ENV_DB_URL) up
	@printf "$(GREEN)[DB]$(NC) Migration has finished.\n\n"

db\:rollback:
	@printf "\n$(RED)[DB]$(NC) Migration rollback has started.\n"
	@docker run -v $(DB_MIGRATE_VOL_MAP) --network $(ROOT_NETWORK) migrate/migrate -verbose -path=$(DB_MIGRATE_PATH) -database $(ENV_DB_URL) down 1
	@printf "$(GREEN)[DB]$(NC) Migration rollback has finished.\n\n"

# --- Migrations
db\:migrate\:create:
	docker run -v $(DB_MIGRATE_VOL_MAP) --network $(ROOT_NETWORK) migrate/migrate create -ext sql -dir $(DB_MIGRATE_PATH) -seq $(name)

db\:migrate\:force:
	docker run -v $(DB_MIGRATE_VOL_MAP) --network $(ROOT_NETWORK) migrate/migrate migrate -path $(DB_MIGRATE_PATH) -database $(ENV_DB_URL) force $(version)
