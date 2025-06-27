
DB_DOCKER_SERVICE_NAME := api-db
DB_DOCKER_CONTAINER_NAME := oullin_db
DB_MIGRATE_SERVICE_NAME := db-migrate

# --- Paths
# Define root paths for clarity. Assume ROOT_PATH is exported or defined.
DB_SEEDER_ROOT_PATH := $(ROOT_PATH)/database/seeder
DB_INFRA_ROOT_PATH := $(ROOT_PATH)/database/infra
DB_INFRA_SSL_PATH := $(DB_INFRA_ROOT_PATH)/ssl
DB_INFRA_SCRIPTS_PATH := $(DB_INFRA_ROOT_PATH)/scripts

# --- SSL Certificate Files
DB_INFRA_SERVER_CRT := $(DB_INFRA_SSL_PATH)/server.crt
DB_INFRA_SERVER_CSR := $(DB_INFRA_SSL_PATH)/server.csr
DB_INFRA_SERVER_KEY := $(DB_INFRA_SSL_PATH)/server.key

db\:sh:
	chmod +x $(DB_INFRA_SCRIPTS_PATH)/healthcheck.sh && \
	chmod +x $(DB_INFRA_SCRIPTS_PATH)/run-migration.sh

db\:up:
	@echo "--> Starting database service..."
	docker compose up $(DB_DOCKER_SERVICE_NAME) -d

db\:down:
	@echo "--> Stopping database service..."
	docker compose stop $(DB_DOCKER_SERVICE_NAME)

db\:logs:
	@echo "--> Tailing logs for $(DB_DOCKER_CONTAINER_NAME)..."
	docker logs -f $(DB_DOCKER_CONTAINER_NAME)

db\:bash:
	@echo "--> Opening a bash shell in $(DB_DOCKER_CONTAINER_NAME)..."
	docker exec -it $(DB_DOCKER_CONTAINER_NAME) bash


# ==============================================================================
# SECURE MIGRATION COMMANDS
# These commands leverage the 'db-migrate' service defined in docker-compose.yml,
# which uses a custom script and Docker Secrets for maximum security.
# ==============================================================================

db\:migrate:
	@printf "\n--> Applying all available 'up' migrations...\n"
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) up
	@printf "--> Migration finished.\n\n"

db\:rollback:
	@printf "\n--> Rolling back the last applied migration...\n"
	# The 'down 1' arguments are passed directly to our secure entrypoint script.
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) down 1
	@printf "--> Migration rollback finished.\n\n"

db\:migrate\:create:
	@echo "--> Creating new migration file named: $(name)"
	# We override the service's default command to use 'create'.
	# The arguments are passed to our secure entrypoint script via "$$@".
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) create -ext sql -dir /migrations -seq $(name)

db\:migrate\:force:
	@printf "\n--> Forcing migration to version $(version)...\n"
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) force $(version)
	@printf "--> Force migration finished.\n\n"


# ==============================================================================
# SETUP & CONVENIENCE COMMANDS
# ==============================================================================

db\:fresh:
	@echo "--> Recreating database from a fresh state (all data will be lost)..."
	make db:delete
	make db:up

db\:delete:
	@echo "--> Stopping services and PERMANENTLY DELETING associated volumes..."
	# The -v flag is crucial here; it removes the named volumes, deleting all data.
	docker compose down -v --remove-orphans

db\:secure:
	@echo "--> Generating new self-signed SSL certificates..."
	rm -f $(DB_INFRA_SERVER_CRT) $(DB_INFRA_SERVER_CSR) $(DB_INFRA_SERVER_KEY)
	openssl genpkey -algorithm RSA -out $(DB_INFRA_SERVER_KEY)
	openssl req -new -key $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CSR) -subj "/CN=oullin-db-ssl"
	openssl x509 -req -days 365 -in $(DB_INFRA_SERVER_CSR) -signkey $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CRT)
	@echo "--> SSL certificates created. The container will set its own key permissions on startup."

db\:seed:
	@echo "--> Running database seeder..."
	# This assumes your Go seeder can connect to the Dockerized database.
	# Ensure your .env file points to the correct DB host and port.
	go run $(DB_SEEDER_ROOT_PATH)/main.go

db\:local:
	@echo "--> Connecting to local PostgreSQL instance..."
	# This command is for connecting to a non-Dockerized local DB, as per your original file.
	# It is kept for convenience if you ever run Postgres outside of Docker.
	cd  $(EN_DB_BIN_DIR) && \
	./psql -h $(ENV_DB_HOST) -U $(ENV_DB_USER_NAME) -d $(ENV_DB_DATABASE_NAME) -p $(ENV_DB_PORT)

