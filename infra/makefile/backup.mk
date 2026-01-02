.PHONY: backup\:create backup\:restore backup\:list backup\:cleanup
.PHONY: backup\:cron\:setup backup\:cron\:remove backup\:cron\:show backup\:help

# --- Backup Configuration
BACKUP_SCRIPTS_PATH := $(ROOT_PATH)/infra/scripts
BACKUP_SCRIPT := $(BACKUP_SCRIPTS_PATH)/db-backup.sh
BACKUP_CRON_SCRIPT := $(BACKUP_SCRIPTS_PATH)/setup-cron-backup.sh
BACKUP_DIR := $(STORAGE_PATH)/backups
BACKUP_RETENTION_DAYS ?= 7

# -------------------------------------------------------------------------------------------------------------------- #
# --- Database Backup Commands
# -------------------------------------------------------------------------------------------------------------------- #

backup\:create:
	@echo -e "$(GREEN)Creating database backup...$(NC)"
	@$(BACKUP_SCRIPT) backup

backup\:restore:
	@if [ -z "$(file)" ]; then \
		echo -e "$(RED)Error: file parameter is required$(NC)"; \
		echo -e "$(YELLOW)Usage: make backup:restore file=<backup-file>$(NC)"; \
		echo -e "$(YELLOW)Example: make backup:restore file=storage/backups/oullin_db_20260102_153045.sql.gz$(NC)"; \
		exit 1; \
	fi
	@echo -e "$(YELLOW)Restoring database from: $(file)$(NC)"
	@$(BACKUP_SCRIPT) restore --file "$(file)"

backup\:list:
	@$(BACKUP_SCRIPT) list

backup\:cleanup:
	@echo -e "$(YELLOW)Cleaning up backups older than $(BACKUP_RETENTION_DAYS) days...$(NC)"
	@BACKUP_RETENTION_DAYS=$(BACKUP_RETENTION_DAYS) $(BACKUP_SCRIPT) cleanup

# -------------------------------------------------------------------------------------------------------------------- #
# --- Automated Backup (Cron) Commands
# -------------------------------------------------------------------------------------------------------------------- #

backup\:cron\:setup:
	@echo -e "$(GREEN)Setting up automated database backups...$(NC)"
	@if [ -n "$(schedule)" ]; then \
		$(BACKUP_CRON_SCRIPT) --schedule "$(schedule)" --retention $(BACKUP_RETENTION_DAYS); \
	else \
		$(BACKUP_CRON_SCRIPT) --retention $(BACKUP_RETENTION_DAYS); \
	fi

backup\:cron\:setup\:dry-run:
	@echo -e "$(YELLOW)Previewing cron setup (dry run)...$(NC)"
	@if [ -n "$(schedule)" ]; then \
		$(BACKUP_CRON_SCRIPT) --schedule "$(schedule)" --retention $(BACKUP_RETENTION_DAYS) --dry-run; \
	else \
		$(BACKUP_CRON_SCRIPT) --retention $(BACKUP_RETENTION_DAYS) --dry-run; \
	fi

backup\:cron\:remove:
	@echo -e "$(YELLOW)Removing automated backup cron jobs...$(NC)"
	@$(BACKUP_CRON_SCRIPT) --remove

backup\:cron\:show:
	@$(BACKUP_CRON_SCRIPT) --show

# -------------------------------------------------------------------------------------------------------------------- #
# --- Backup Help
# -------------------------------------------------------------------------------------------------------------------- #

backup\:help:
	@echo -e "\n$(BOLD)$(CYAN)Database Backup Commands$(NC)\n"
	@echo -e "$(BOLD)$(BLUE)Backup Operations:$(NC)"
	@echo -e "  $(BOLD)$(GREEN)backup:create$(NC)              : Create a new database backup"
	@echo -e "  $(BOLD)$(GREEN)backup:restore$(NC)             : Restore database from a backup file"
	@echo -e "                                  $(YELLOW)Usage: make backup:restore file=<backup-file>$(NC)"
	@echo -e "  $(BOLD)$(GREEN)backup:list$(NC)                : List all available backups"
	@echo -e "  $(BOLD)$(GREEN)backup:cleanup$(NC)             : Remove backups older than retention period"
	@echo -e "                                  $(YELLOW)Default: $(BACKUP_RETENTION_DAYS) days$(NC)"
	@echo -e ""
	@echo -e "$(BOLD)$(BLUE)Automated Backups (Cron):$(NC)"
	@echo -e "  $(BOLD)$(GREEN)backup:cron:setup$(NC)          : Setup automated weekly backups (Sundays at 2 AM)"
	@echo -e "                                  $(YELLOW)Usage: make backup:cron:setup [schedule=\"0 2 * * 0\"]$(NC)"
	@echo -e "  $(BOLD)$(GREEN)backup:cron:setup:dry-run$(NC)  : Preview cron setup without applying"
	@echo -e "  $(BOLD)$(GREEN)backup:cron:remove$(NC)         : Remove automated backup cron jobs"
	@echo -e "  $(BOLD)$(GREEN)backup:cron:show$(NC)           : Show current backup cron jobs"
	@echo -e ""
	@echo -e "$(BOLD)$(BLUE)Examples:$(NC)"
	@echo -e "  $(YELLOW)# Create a backup$(NC)"
	@echo -e "  make backup:create"
	@echo -e ""
	@echo -e "  $(YELLOW)# Restore from a backup$(NC)"
	@echo -e "  make backup:restore file=storage/backups/oullin_db_20260102_153045.sql.gz"
	@echo -e ""
	@echo -e "  $(YELLOW)# Setup weekly backups (Sundays at 2 AM)$(NC)"
	@echo -e "  make backup:cron:setup"
	@echo -e ""
	@echo -e "  $(YELLOW)# Setup daily backups at 2 AM$(NC)"
	@echo -e "  make backup:cron:setup schedule=\"0 2 * * *\""
	@echo -e ""
	@echo -e "  $(YELLOW)# Clean up backups older than 14 days$(NC)"
	@echo -e "  make BACKUP_RETENTION_DAYS=14 backup:cleanup"
	@echo -e ""
