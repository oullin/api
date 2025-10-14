.PHONY: help
.EXPORT_ALL_VARIABLES:

# -------------------------------------------------------------------------------------------------------------------- #
# -------------------------------------------------------------------------------------------------------------------- #

SHELL := /bin/bash

# -------------------------------------------------------------------------------------------------------------------- #
# -------------------------------------------------------------------------------------------------------------------- #

NC     := \033[0m
BOLD   := \033[1m
CYAN   := \033[36m
WHITE  := \033[37m
GREEN  := \033[0;32m
BLUE   := \033[0;34m
RED    := \033[0;31m
YELLOW := \033[1;33m

# -------------------------------------------------------------------------------------------------------------------- #
# -------------------------------------------------------------------------------------------------------------------- #

ROOT_NETWORK          := oullin_net
DATABASE              := "api-db"
SOURCE                := go_bindata
ROOT_PATH             := $(shell pwd)
APP_PATH              := $(ROOT_PATH)/
STORAGE_PATH          := $(ROOT_PATH)/storage
REPO_OWNER            := $(shell cd .. && basename "$$(pwd)")
VERSION               := $(shell git describe --tags 2>/dev/null | cut -c 2-)
CGO_ENABLED           := 1

# -------------------------------------------------------------------------------------------------------------------- #
# -------------------------------------------------------------------------------------------------------------------- #

include ./metal/makefile/helpers.mk
include ./metal/makefile/env.mk
include ./metal/makefile/db.mk
include ./metal/makefile/app.mk
include ./metal/makefile/logs.mk
include ./metal/makefile/build.mk
include ./metal/makefile/infra.mk
include ./metal/makefile/caddy.mk

# -------------------------------------------------------------------------------------------------------------------- #
# -------------------------------------------------------------------------------------------------------------------- #

help:
	@printf "\n$(BOLD)$(CYAN)Applications Options$(NC)\n"
	@printf "$(WHITE)Usage:$(NC) make $(BOLD)$(YELLOW)<target>$(NC)\n\n"

	@printf "$(BOLD)$(BLUE)General Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)fresh$(NC)            : Clean and reset various project components (logs, build, etc.).\n"
	@printf "  $(BOLD)$(GREEN)audit$(NC)            : Run code audits and checks.\n"
	@printf "  $(BOLD)$(GREEN)watch$(NC)            : Start a file watcher process.\n"
	@printf "  $(BOLD)$(GREEN)format$(NC)           : Automatically format code.\n"
	@printf "  $(BOLD)$(GREEN)test-all$(NC)         : Run all the application tests.\n"
	@printf "  $(BOLD)$(GREEN)run-cli$(NC)          : Run the application CLI interface.\n"
	@printf "  $(BOLD)$(GREEN)run-cli-docker$(NC)   : Run the application [docker] dev's CLI interface.\n\n"
	@printf "  $(BOLD)$(GREEN)run-metal$(NC)        : Run the application dev's CLI interface.\n\n"

	@printf "$(BOLD)$(BLUE)Build Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)build-local$(NC)      : Build the main application for development.\n"
	@printf "  $(BOLD)$(GREEN)build-ci$(NC)         : Build the main application for the CI.\n"
	@printf "  $(BOLD)$(GREEN)build-release$(NC)    : Build a release version of the application.\n"
	@printf "  $(BOLD)$(GREEN)build-fresh$(NC)      : Build a fresh development environment.\n\n"

	@printf "$(BOLD)$(BLUE)Database Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)db:local$(NC)         : Set up or manage the local database environment.\n"
	@printf "  $(BOLD)$(GREEN)db:up$(NC)            : Start the database service or container.\n"
	@printf "  $(BOLD)$(GREEN)db:ping$(NC)          : Check the database connection.\n"
	@printf "  $(BOLD)$(GREEN)db:bash$(NC)          : Access the database environment via bash.\n"
	@printf "  $(BOLD)$(GREEN)db:fresh$(NC)         : Reset and re-seed the database.\n"
	@printf "  $(BOLD)$(GREEN)db:logs$(NC)          : View database logs.\n"
	@printf "  $(BOLD)$(GREEN)db:delete$(NC)        : Delete the database.\n"
	@printf "  $(BOLD)$(GREEN)db:secure$(NC)        : Apply database security configurations.\n"
	@printf "  $(BOLD)$(GREEN)db:secure:show$(NC)   : Display database security configurations.\n"
	@printf "  $(BOLD)$(GREEN)db:chmod$(NC)         : Adjust database file or directory permissions.\n"
	@printf "  $(BOLD)$(GREEN)db:seed$(NC)          : Run database seeders to populate data.\n"
	@printf "  $(BOLD)$(GREEN)db:import$(NC)        : Execute SQL statements from ./storage/sql/dump.sql.\n"
	@printf "  $(BOLD)$(GREEN)db:migrate$(NC)       : Run database migrations.\n"
	@printf "  $(BOLD)$(GREEN)db:rollback$(NC)      : Rollback database migrations (usually the last batch).\n"
	@printf "  $(BOLD)$(GREEN)db:migrate:create$(NC): Create a new database migration file.\n"
	@printf "  $(BOLD)$(GREEN)db:migrate:force$(NC) : Force database migrations to run.\n\n"

	@printf "$(BOLD)$(BLUE)Environment Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)env:check$(NC)        : Verify environment configuration.\n"
	@printf "  $(BOLD)$(GREEN)env:fresh$(NC)        : Refresh environment settings.\n"
	@printf "  $(BOLD)$(GREEN)env:init$(NC)         : Initialize environment settings.\n"
	@printf "  $(BOLD)$(GREEN)env:print$(NC)        : Display current environment settings.\n\n"

	@printf "$(BOLD)$(BLUE)Log Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)logs:fresh$(NC)       : Clear application logs.\n\n"

	@printf "$(BOLD)$(BLUE)Infrastructure Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)supv:api:status$(NC)  : Shows the API service supervisor state.\n"
	@printf "  $(BOLD)$(GREEN)supv:api:start$(NC)   : Start the API service supervisor.\n"
	@printf "  $(BOLD)$(GREEN)supv:api:stop$(NC)    : Stop the API service supervisor.\n"
	@printf "  $(BOLD)$(GREEN)supv:api:restart$(NC) : Restart the API service supervisor.\n"
	@printf "  $(BOLD)$(GREEN)supv:api:logs$(NC)    : Show the the API service supervisor logs.\n"
	@printf "  $(BOLD)$(GREEN)supv:api:logs-err$(NC): Show the the API service supervisor error logs.\n\n"

	@printf "$(BOLD)$(BLUE)Caddy Commands:$(NC)\n"
	@printf "  $(BOLD)$(GREEN)caddy-gen-cert$(NC)   : Generate the caddy's mtls certificates.\n"
	@printf "  $(BOLD)$(GREEN)caddy-del-cert$(NC)   : Remove the caddy's mtls certificates.\n"
	@printf "  $(BOLD)$(GREEN)caddy-validate$(NC)   : Validates caddy's files syntax.\n"

	@printf "$(NC)\n"
