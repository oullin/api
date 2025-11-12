# -------------------------------------------------------------------------------------------------------------------- #
# Monitoring Stack Targets
# -------------------------------------------------------------------------------------------------------------------- #

# -------------------------------------------------------------------------------------------------------------------- #
# Configuration Variables
# -------------------------------------------------------------------------------------------------------------------- #

ROOT_PATH           := $(shell pwd)
MONITORING_DIR      := $(ROOT_PATH)/infra/metrics
BACKUPS_DIR         := $(ROOT_PATH)/storage/monitoring/backups

# Monitoring service URLs and ports
GRAFANA_HOST        := localhost
GRAFANA_PORT        := 3000
GRAFANA_URL         := http://$(GRAFANA_HOST):$(GRAFANA_PORT)

PROMETHEUS_HOST     := localhost
PROMETHEUS_PORT     := 9090
PROMETHEUS_URL      := http://$(PROMETHEUS_HOST):$(PROMETHEUS_PORT)

CADDY_ADMIN_HOST    := localhost
CADDY_ADMIN_PORT    := 2019
CADDY_ADMIN_URL     := http://$(CADDY_ADMIN_HOST):$(CADDY_ADMIN_PORT)

API_HOST            := localhost
API_PORT            := 18080
API_URL             := http://$(API_HOST):$(API_PORT)
PING_USERNAME       ?= $(ENV_PING_USERNAME)
PING_PASSWORD       ?= $(ENV_PING_PASSWORD)
PING_AUTH_FLAG      := $(if $(and $(PING_USERNAME),$(PING_PASSWORD)),-u $(PING_USERNAME):$(PING_PASSWORD),)

# Production API endpoint (behind Caddy)
API_PROD_HOST       := localhost
API_PROD_URL        := http://$(API_PROD_HOST)

# Internal service URLs (Docker network)
PG_EXPORTER_HOST    := postgres_exporter_local
PG_EXPORTER_PORT    := 9187
PG_EXPORTER_URL     := http://$(PG_EXPORTER_HOST):$(PG_EXPORTER_PORT)

# -------------------------------------------------------------------------------------------------------------------- #
# PHONY Targets
# -------------------------------------------------------------------------------------------------------------------- #

.PHONY: monitor-up monitor-up-prod monitor-down monitor-down-prod monitor-restart monitor-restart-prod \
	monitor-up-full monitor-up-full-prod monitor-up-logs monitor-up-logs-prod monitor-down-remove monitor-down-remove-prod \
	monitor-pull monitor-pull-prod monitor-docker-config monitor-docker-config-prod monitor-docker-exec-prometheus monitor-docker-exec-prometheus-prod \
	monitor-docker-exec-grafana monitor-docker-exec-grafana-prod monitor-docker-ps monitor-docker-inspect monitor-docker-inspect-prod \
	monitor-docker-logs-prometheus monitor-docker-logs-prometheus-prod monitor-docker-logs-grafana monitor-docker-logs-grafana-prod monitor-docker-logs-db monitor-docker-logs-db-prod \
	monitor-status monitor-logs monitor-logs-prod \
	monitor-test monitor-targets monitor-config monitor-config-prod monitor-grafana monitor-prometheus \
	monitor-caddy-metrics monitor-api-metrics monitor-db-metrics monitor-db-metrics-prod monitor-metrics \
	monitor-traffic monitor-traffic-heavy monitor-traffic-prod monitor-traffic-heavy-prod \
	monitor-clean monitor-clean-prod monitor-stats monitor-stats-prod monitor-backup monitor-backup-prod monitor-export-dashboards monitor-help

# -------------------------------------------------------------------------------------------------------------------- #
# Start/Stop Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Start monitoring stack (local development)
monitor-up:
	@printf "$(BOLD)$(CYAN)Starting monitoring stack (local)...$(NC)\n"
	@docker compose --profile local up -d prometheus_local grafana_local postgres_exporter_local
	@sleep 3
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack started$(NC)\n"
	@printf "\n$(BOLD)Access points:$(NC)\n"
	@printf "  $(GREEN)Grafana:$(NC)     $(GRAFANA_URL)\n"
	@printf "  $(GREEN)Prometheus:$(NC)  $(PROMETHEUS_URL)\n"
	@printf "  $(GREEN)Caddy Admin:$(NC) $(CADDY_ADMIN_URL)\n\n"

## Start monitoring stack (production)
monitor-up-prod:
	@printf "$(BOLD)$(CYAN)Starting monitoring stack (production)...$(NC)\n"
	@docker compose --profile prod up -d prometheus grafana postgres_exporter
	@sleep 3
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack started$(NC)\n"
	@printf "\n$(BOLD)Access points (from server):$(NC)\n"
	@printf "  $(GREEN)Grafana:$(NC)     $(GRAFANA_URL)\n"
	@printf "  $(GREEN)Prometheus:$(NC)  $(PROMETHEUS_URL)\n"
	@printf "  $(GREEN)Caddy Admin:$(NC) $(CADDY_ADMIN_URL)\n\n"

## Stop monitoring stack (local)
monitor-down:
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack (local)...$(NC)\n"
	@docker compose --profile local stop prometheus_local grafana_local postgres_exporter_local
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack stopped$(NC)\n\n"

## Stop monitoring stack (production)
monitor-down-prod:
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack (production)...$(NC)\n"
	@docker compose --profile prod stop prometheus grafana postgres_exporter
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack stopped$(NC)\n\n"

## Restart monitoring stack (local)
monitor-restart:
	@printf "$(BOLD)$(CYAN)Restarting monitoring stack (local)...$(NC)\n"
	@docker compose --profile local restart prometheus_local grafana_local postgres_exporter_local
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack restarted$(NC)\n\n"

## Restart monitoring stack (production)
monitor-restart-prod:
	@printf "$(BOLD)$(CYAN)Restarting monitoring stack (production)...$(NC)\n"
	@docker compose --profile prod restart prometheus grafana postgres_exporter
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack restarted$(NC)\n\n"

# -------------------------------------------------------------------------------------------------------------------- #
# Docker Compose Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Start monitoring with full stack (API + DB + monitoring) - local
monitor-up-full:
	@printf "$(BOLD)$(CYAN)Starting full stack with monitoring (local)...$(NC)\n"
	@docker compose --profile local up -d
	@sleep 3
	@printf "$(BOLD)$(GREEN)✓ Full stack started$(NC)\n\n"

## Start monitoring with full stack (API + DB + monitoring) - production
monitor-up-full-prod:
	@printf "$(BOLD)$(CYAN)Starting full stack with monitoring (production)...$(NC)\n"
	@docker compose --profile prod up -d
	@sleep 3
	@printf "$(BOLD)$(GREEN)✓ Full stack started$(NC)\n\n"

## Start monitoring stack with logs (foreground) - local
monitor-up-logs:
	@printf "$(BOLD)$(CYAN)Starting monitoring stack with logs (local)...$(NC)\n"
	@docker compose --profile local up prometheus_local grafana_local postgres_exporter_local

## Start monitoring stack with logs (foreground) - production
monitor-up-logs-prod:
	@printf "$(BOLD)$(CYAN)Starting monitoring stack with logs (production)...$(NC)\n"
	@docker compose --profile prod up prometheus grafana postgres_exporter

## Stop and remove monitoring containers - local
monitor-down-remove:
	@printf "$(BOLD)$(CYAN)Stopping and removing monitoring containers (local)...$(NC)\n"
	@docker compose --profile local down prometheus_local grafana_local postgres_exporter_local
	@printf "$(BOLD)$(GREEN)✓ Containers stopped and removed$(NC)\n\n"

## Stop and remove monitoring containers - production
monitor-down-remove-prod:
	@printf "$(BOLD)$(CYAN)Stopping and removing monitoring containers (production)...$(NC)\n"
	@docker compose --profile prod down prometheus grafana postgres_exporter
	@printf "$(BOLD)$(GREEN)✓ Containers stopped and removed$(NC)\n\n"

## Pull latest monitoring images (local)
monitor-pull:
	@printf "$(BOLD)$(CYAN)Pulling latest monitoring images (local)...$(NC)\n"
	@docker compose pull prometheus_local grafana_local postgres_exporter_local
	@printf "$(BOLD)$(GREEN)✓ Images pulled$(NC)\n\n"

## Pull latest monitoring images (production)
monitor-pull-prod:
	@printf "$(BOLD)$(CYAN)Pulling latest monitoring images (production)...$(NC)\n"
	@docker compose pull prometheus grafana postgres_exporter
	@printf "$(BOLD)$(GREEN)✓ Images pulled$(NC)\n\n"

## Show docker compose config for monitoring services (local)
monitor-docker-config:
	@printf "$(BOLD)$(CYAN)Docker Compose Configuration (monitoring - local)$(NC)\n\n"
	@docker compose config --profile local | grep -A 20 "prometheus_local\|grafana_local\|postgres_exporter_local" || docker compose config --profile local

## Show docker compose config for monitoring services (production)
monitor-docker-config-prod:
	@printf "$(BOLD)$(CYAN)Docker Compose Configuration (monitoring - production)$(NC)\n\n"
	@docker compose config --profile prod | grep -A 20 "prometheus\|grafana\|postgres_exporter" || docker compose config --profile prod

## Execute command in Prometheus container (local)
monitor-docker-exec-prometheus:
	@printf "$(BOLD)$(CYAN)Executing shell in Prometheus container (local)...$(NC)\n"
	@docker exec -it oullin_prometheus_local /bin/sh

## Execute command in Prometheus container (production)
monitor-docker-exec-prometheus-prod:
	@printf "$(BOLD)$(CYAN)Executing shell in Prometheus container (production)...$(NC)\n"
	@docker exec -it oullin_prometheus /bin/sh

## Execute command in Grafana container (local)
monitor-docker-exec-grafana:
	@printf "$(BOLD)$(CYAN)Executing shell in Grafana container (local)...$(NC)\n"
	@docker exec -it oullin_grafana_local /bin/sh

## Execute command in Grafana container (production)
monitor-docker-exec-grafana-prod:
	@printf "$(BOLD)$(CYAN)Executing shell in Grafana container (production)...$(NC)\n"
	@docker exec -it oullin_grafana /bin/sh

## Show docker ps for monitoring containers
monitor-docker-ps:
	@printf "$(BOLD)$(CYAN)Monitoring Containers$(NC)\n\n"
	@docker ps --filter "name=prometheus" --filter "name=grafana" --filter "name=exporter" --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}"
	@printf "\n"

## Show docker inspect for monitoring containers (local)
monitor-docker-inspect:
	@printf "$(BOLD)$(CYAN)Inspecting Monitoring Containers (local)$(NC)\n\n"
	@docker inspect oullin_prometheus_local oullin_grafana_local oullin_postgres_exporter_local 2>/dev/null | jq '.[].Name, .[].State, .[].NetworkSettings.Networks' || echo "$(RED)Containers not running$(NC)"

## Show docker inspect for monitoring containers (production)
monitor-docker-inspect-prod:
	@printf "$(BOLD)$(CYAN)Inspecting Monitoring Containers (production)$(NC)\n\n"
	@docker inspect oullin_prometheus oullin_grafana oullin_postgres_exporter 2>/dev/null | jq '.[].Name, .[].State, .[].NetworkSettings.Networks' || echo "$(RED)Containers not running$(NC)"

## View monitoring container logs (docker logs - local)
monitor-docker-logs-prometheus:
	@docker logs -f oullin_prometheus_local

monitor-docker-logs-grafana:
	@docker logs -f oullin_grafana_local

monitor-docker-logs-db:
	@docker logs -f oullin_postgres_exporter_local

## View monitoring container logs (docker logs - production)
monitor-docker-logs-prometheus-prod:
	@docker logs -f oullin_prometheus

monitor-docker-logs-grafana-prod:
	@docker logs -f oullin_grafana

monitor-docker-logs-db-prod:
	@docker logs -f oullin_postgres_exporter

# -------------------------------------------------------------------------------------------------------------------- #
# Status & Information Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Show status of monitoring services
monitor-status:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Status$(NC)\n\n"
	@docker ps --filter "name=prometheus" --filter "name=grafana" --filter "name=exporter" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@printf "\n"

## Show logs from all monitoring services (local)
monitor-logs:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Logs (local)$(NC)\n\n"
	@docker compose logs -f prometheus_local grafana_local postgres_exporter_local

## Show logs from all monitoring services (production)
monitor-logs-prod:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Logs (production)$(NC)\n\n"
	@docker compose logs -f prometheus grafana postgres_exporter

# -------------------------------------------------------------------------------------------------------------------- #
# Testing & Verification Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Run full monitoring stack test suite (local profile only)
monitor-test:
	@printf "$(BOLD)$(CYAN)Running monitoring stack tests (local profile)...$(NC)\n"
	@printf "$(YELLOW)Note: This target is for local development only.$(NC)\n"
	@printf "$(YELLOW)For production, verify monitoring from the server directly.$(NC)\n\n"
	@printf "$(BOLD)1. Checking services are running...$(NC)\n"
	@docker ps --filter "name=prometheus_local" --filter "name=grafana_local" --filter "name=postgres_exporter_local" --format "  ✓ {{.Names}}: {{.Status}}" || echo "  $(RED)✗ Services not running$(NC)"
	@printf "\n$(BOLD)2. Testing Prometheus targets...$(NC)\n"
	@curl -s $(PROMETHEUS_URL)/api/v1/targets | grep -q '"health":"up"' && echo "  $(GREEN)✓ Prometheus targets are UP$(NC)" || echo "  $(RED)✗ Some targets are DOWN$(NC)"
	@printf "\n$(BOLD)3. Testing Caddy metrics endpoint...$(NC)\n"
	@curl -s $(CADDY_ADMIN_URL)/metrics | grep -q "caddy_http_requests_total" && echo "  $(GREEN)✓ Caddy metrics accessible$(NC)" || echo "  $(RED)✗ Caddy metrics unavailable$(NC)"
	@printf "\n$(BOLD)4. Testing API metrics endpoint...$(NC)\n"
	@curl -s $(API_URL)/metrics | grep -q "go_goroutines" && echo "  $(GREEN)✓ API metrics accessible$(NC)" || echo "  $(RED)✗ API metrics unavailable$(NC)"
	@printf "\n$(BOLD)5. Testing Grafana...$(NC)\n"
	@curl -s $(GRAFANA_URL)/api/health | grep -q "ok" && echo "  $(GREEN)✓ Grafana is healthy$(NC)" || echo "  $(RED)✗ Grafana is unhealthy$(NC)"
	@printf "\n$(BOLD)$(GREEN)Test suite completed!$(NC)\n\n"

## Verify Prometheus targets status
monitor-targets:
	@printf "$(BOLD)$(CYAN)Prometheus Targets Status$(NC)\n\n"
	@curl -s $(PROMETHEUS_URL)/api/v1/targets | jq -r '.data.activeTargets[] | "[\(.health | ascii_upcase)] \(.labels.job) - \(.scrapeUrl)"' || echo "$(RED)Failed to fetch targets. Is Prometheus running?$(NC)"
	@printf "\n"

## Check Prometheus configuration (local)
monitor-config:
	@printf "$(BOLD)$(CYAN)Prometheus Configuration (local)$(NC)\n\n"
	@docker exec oullin_prometheus_local cat /etc/prometheus/prometheus.yml

## Check Prometheus configuration (production)
monitor-config-prod:
	@printf "$(BOLD)$(CYAN)Prometheus Configuration (production)$(NC)\n\n"
	@docker exec oullin_prometheus cat /etc/prometheus/prometheus.yml

# -------------------------------------------------------------------------------------------------------------------- #
# Metrics Access Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Open Grafana in browser
monitor-grafana:
	@printf "$(BOLD)$(CYAN)Opening Grafana...$(NC)\n"
	@printf "URL: $(GREEN)$(GRAFANA_URL)$(NC)\n"
	@printf "Credentials: admin / (set via GRAFANA_ADMIN_PASSWORD)\n\n"
	@which xdg-open > /dev/null && xdg-open $(GRAFANA_URL) || which open > /dev/null && open $(GRAFANA_URL) || echo "Please open $(GRAFANA_URL) in your browser"

## Open Prometheus in browser
monitor-prometheus:
	@printf "$(BOLD)$(CYAN)Opening Prometheus...$(NC)\n"
	@printf "URL: $(GREEN)$(PROMETHEUS_URL)$(NC)\n\n"
	@which xdg-open > /dev/null && xdg-open $(PROMETHEUS_URL) || which open > /dev/null && open $(PROMETHEUS_URL) || echo "Please open $(PROMETHEUS_URL) in your browser"

## Show Caddy metrics
monitor-caddy-metrics:
	@printf "$(BOLD)$(CYAN)Caddy Metrics$(NC)\n\n"
	@curl -s $(CADDY_ADMIN_URL)/metrics | grep "^caddy_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n"
	@printf "Full metrics: $(GREEN)$(CADDY_ADMIN_URL)/metrics$(NC)\n\n"

## Show API metrics
monitor-api-metrics:
	@printf "$(BOLD)$(CYAN)API Metrics$(NC)\n\n"
	@curl -s $(API_URL)/metrics | grep "^go_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n"
	@printf "Full metrics: $(GREEN)$(API_URL)/metrics$(NC)\n\n"

## Show PostgreSQL metrics (local)
monitor-db-metrics:
	@printf "$(BOLD)$(CYAN)PostgreSQL Metrics (local)$(NC)\n\n"
	@docker exec oullin_prometheus_local curl -s $(PG_EXPORTER_URL)/metrics | grep "^pg_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n\n"

## Show PostgreSQL metrics (production)
monitor-db-metrics-prod:
	@printf "$(BOLD)$(CYAN)PostgreSQL Metrics (production)$(NC)\n\n"
	@docker exec oullin_prometheus curl -s http://postgres_exporter:9187/metrics | grep "^pg_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n\n"

## Show all metrics endpoints
monitor-metrics:
	@printf "$(BOLD)$(CYAN)Available Metrics Endpoints$(NC)\n\n"
	@printf "  $(GREEN)Caddy:$(NC)      $(CADDY_ADMIN_URL)/metrics\n"
	@printf "  $(GREEN)API:$(NC)        $(API_URL)/metrics\n"
	@printf "  $(GREEN)PostgreSQL:$(NC) $(PG_EXPORTER_URL)/metrics (internal)\n"
	@printf "  $(GREEN)Prometheus:$(NC) $(PROMETHEUS_URL)/metrics\n\n"

# -------------------------------------------------------------------------------------------------------------------- #
# Traffic Generation & Testing
# -------------------------------------------------------------------------------------------------------------------- #

## Generate test traffic to populate metrics (local profile)
monitor-traffic:
	@if [ -z "$(PING_USERNAME)" ] || [ -z "$(PING_PASSWORD)" ]; then \
		printf "$(RED)Missing ping credentials. Export ENV_PING_USERNAME/ENV_PING_PASSWORD or pass PING_USERNAME/PING_PASSWORD to make.$(NC)\n"; \
		exit 1; \
	fi
	@printf "$(BOLD)$(CYAN)Generating test traffic (local)...$(NC)\n"
	@printf "Making 100 requests to /ping endpoint...\n"
	@for i in $$(seq 1 100); do \
		curl -s $(PING_AUTH_FLAG) $(API_URL)/ping > /dev/null && printf "." || printf "$(RED)✗$(NC)"; \
		sleep 0.1; \
	done
	@printf "\n$(BOLD)$(GREEN)✓ Test traffic generated$(NC)\n"
	@printf "\nCheck dashboards at: $(GREEN)$(GRAFANA_URL)$(NC)\n\n"

## Generate heavy test traffic (local profile)
monitor-traffic-heavy:
	@printf "$(BOLD)$(CYAN)Generating heavy test traffic (local)...$(NC)\n"
	@printf "Making 500 requests with 5 concurrent connections...\n"
	@for i in $$(seq 1 100); do \
		(for j in $$(seq 1 5); do curl -s $(API_URL)/ping > /dev/null & done; wait); \
		printf "."; \
		sleep 0.05; \
	done
	@printf "\n$(BOLD)$(GREEN)✓ Heavy test traffic generated$(NC)\n\n"

## Generate test traffic to populate metrics (production profile)
monitor-traffic-prod:
	@printf "$(BOLD)$(CYAN)Generating test traffic (production)...$(NC)\n"
	@printf "Making 100 requests to /api/ping endpoint...\n"
	@for i in $$(seq 1 100); do \
		curl -s $(API_PROD_URL)/api/ping > /dev/null && printf "." || printf "$(RED)✗$(NC)"; \
		sleep 0.1; \
	done
	@printf "\n$(BOLD)$(GREEN)✓ Test traffic generated$(NC)\n"
	@printf "\n$(YELLOW)Note: Run this from the production server$(NC)\n"
	@printf "SSH tunnel for Grafana: $(GREEN)ssh -L 3000:localhost:3000 user@server$(NC)\n\n"

## Generate heavy test traffic (production profile)
monitor-traffic-heavy-prod:
	@printf "$(BOLD)$(CYAN)Generating heavy test traffic (production)...$(NC)\n"
	@printf "Making 500 requests with 5 concurrent connections...\n"
	@for i in $$(seq 1 100); do \
		(for j in $$(seq 1 5); do curl -s $(API_PROD_URL)/api/ping > /dev/null & done; wait); \
		printf "."; \
		sleep 0.05; \
	done
	@printf "\n$(BOLD)$(GREEN)✓ Heavy test traffic generated$(NC)\n"
	@printf "\n$(YELLOW)Note: Run this from the production server$(NC)\n\n"

# -------------------------------------------------------------------------------------------------------------------- #
# Utility Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Clean monitoring data (removes all metrics/dashboard data) - local
monitor-clean:
	@printf "$(BOLD)$(RED)WARNING: This will delete all monitoring data (local)!$(NC)\n"
	@printf "Press Ctrl+C to cancel, or Enter to continue..."
	@read
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack...$(NC)\n"
	@docker compose --profile local down prometheus_local grafana_local
	@printf "$(BOLD)$(CYAN)Removing volumes...$(NC)\n"
	@docker volume rm -f prometheus_data grafana_data || true
	@printf "$(BOLD)$(GREEN)✓ Monitoring data cleaned$(NC)\n\n"

## Clean monitoring data (removes all metrics/dashboard data) - production
monitor-clean-prod:
	@printf "$(BOLD)$(RED)WARNING: This will delete all monitoring data (production)!$(NC)\n"
	@printf "Press Ctrl+C to cancel, or Enter to continue..."
	@read
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack...$(NC)\n"
	@docker compose --profile prod down prometheus grafana
	@printf "$(BOLD)$(CYAN)Removing volumes...$(NC)\n"
	@docker volume rm -f prometheus_prod_data grafana_prod_data || true
	@printf "$(BOLD)$(GREEN)✓ Monitoring data cleaned$(NC)\n\n"

## Show monitoring stack resource usage (local)
monitor-stats:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Resource Usage (local)$(NC)\n\n"
	@docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" \
		oullin_prometheus_local oullin_grafana_local oullin_postgres_exporter_local 2>/dev/null || \
		echo "$(RED)No monitoring containers running$(NC)"
	@printf "\n"

## Show monitoring stack resource usage (production)
monitor-stats-prod:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Resource Usage (production)$(NC)\n\n"
	@docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" \
		oullin_prometheus oullin_grafana oullin_postgres_exporter 2>/dev/null || \
		echo "$(RED)No monitoring containers running$(NC)"
	@printf "\n"

## Backup Prometheus data (with automatic rotation) - local
monitor-backup:
	@printf "$(BOLD)$(CYAN)Backing up Prometheus data (local)...$(NC)\n"
	@mkdir -p $(BACKUPS_DIR)
	@docker run --rm -v prometheus_data:/data -v $(PWD)/backups:/backup alpine \
		tar czf /backup/prometheus-backup-$$(date +%Y%m%d-%H%M%S).tar.gz /data
	@printf "$(BOLD)$(GREEN)✓ Backup created in $(BACKUPS_DIR)/$(NC)\n"
	@printf "$(YELLOW)Rotating backups (keeping last 5)...$(NC)\n"
	@for f in $$(ls -t $(BACKUPS_DIR)/prometheus-backup-*.tar.gz 2>/dev/null | tail -n +6); do rm -f "$$f"; done || true
	@BACKUP_COUNT=$$(ls -1 $(BACKUPS_DIR)/prometheus-backup-*.tar.gz 2>/dev/null | wc -l); \
		printf "$(BOLD)$(GREEN)✓ Backup rotation complete ($${BACKUP_COUNT} backups kept)$(NC)\n\n"

## Backup Prometheus data (with automatic rotation) - production
monitor-backup-prod:
	@printf "$(BOLD)$(CYAN)Backing up Prometheus data (production)...$(NC)\n"
	@mkdir -p $(BACKUPS_DIR)
	@docker run --rm -v prometheus_prod_data:/data -v $(PWD)/backups:/backup alpine \
		tar czf /backup/prometheus-prod-backup-$$(date +%Y%m%d-%H%M%S).tar.gz /data
	@printf "$(BOLD)$(GREEN)✓ Backup created in $(BACKUPS_DIR)/$(NC)\n"
	@printf "$(YELLOW)Rotating backups (keeping last 5)...$(NC)\n"
	@for f in $$(ls -t $(BACKUPS_DIR)/prometheus-prod-backup-*.tar.gz 2>/dev/null | tail -n +6); do rm -f "$$f"; done || true
	@BACKUP_COUNT=$$(ls -1 $(BACKUPS_DIR)/prometheus-prod-backup-*.tar.gz 2>/dev/null | wc -l); \
		printf "$(BOLD)$(GREEN)✓ Backup rotation complete ($${BACKUP_COUNT} backups kept)$(NC)\n\n"

## Export Grafana dashboards to JSON files
monitor-export-dashboards:
	@printf "$(BOLD)$(CYAN)Exporting Grafana dashboards...$(NC)\n"
	@$(MONITORING_DIR)/grafana/scripts/export-dashboards.sh

## Show monitoring help
monitor-help:
	@printf "\n$(BOLD)$(CYAN)Monitoring Stack Commands$(NC)\n\n"
	@printf "$(BOLD)$(BLUE)Start/Stop:$(NC)\n"
	@printf "  $(GREEN)monitor-up$(NC)                         - Start monitoring stack (local)\n"
	@printf "  $(GREEN)monitor-up-prod$(NC)                    - Start monitoring stack (production)\n"
	@printf "  $(GREEN)monitor-up-full$(NC)                    - Start full stack with monitoring (local)\n"
	@printf "  $(GREEN)monitor-up-full-prod$(NC)               - Start full stack with monitoring (prod)\n"
	@printf "  $(GREEN)monitor-up-logs$(NC)                    - Start with logs in foreground (local)\n"
	@printf "  $(GREEN)monitor-up-logs-prod$(NC)               - Start with logs in foreground (prod)\n"
	@printf "  $(GREEN)monitor-down$(NC)                       - Stop monitoring stack (local)\n"
	@printf "  $(GREEN)monitor-down-prod$(NC)                  - Stop monitoring stack (production)\n"
	@printf "  $(GREEN)monitor-down-remove$(NC)                - Stop and remove containers (local)\n"
	@printf "  $(GREEN)monitor-down-remove-prod$(NC)           - Stop and remove containers (prod)\n"
	@printf "  $(GREEN)monitor-restart$(NC)                    - Restart monitoring stack (local)\n"
	@printf "  $(GREEN)monitor-restart-prod$(NC)               - Restart monitoring stack (prod)\n\n"
	@printf "$(BOLD)$(BLUE)Docker Commands:$(NC)\n"
	@printf "  $(GREEN)monitor-docker-ps$(NC)                  - Show running monitoring containers\n"
	@printf "  $(GREEN)monitor-docker-config$(NC)              - Show docker compose config (local)\n"
	@printf "  $(GREEN)monitor-docker-config-prod$(NC)         - Show docker compose config (prod)\n"
	@printf "  $(GREEN)monitor-docker-inspect$(NC)             - Inspect monitoring containers (local)\n"
	@printf "  $(GREEN)monitor-docker-inspect-prod$(NC)        - Inspect monitoring containers (prod)\n"
	@printf "  $(GREEN)monitor-docker-exec-prometheus$(NC)     - Shell into Prometheus container (local)\n"
	@printf "  $(GREEN)monitor-docker-exec-prometheus-prod$(NC)- Shell into Prometheus container (prod)\n"
	@printf "  $(GREEN)monitor-docker-exec-grafana$(NC)        - Shell into Grafana container (local)\n"
	@printf "  $(GREEN)monitor-docker-exec-grafana-prod$(NC)   - Shell into Grafana container (prod)\n"
	@printf "  $(GREEN)monitor-docker-logs-prometheus$(NC)     - Docker logs for Prometheus (local)\n"
	@printf "  $(GREEN)monitor-docker-logs-prometheus-prod$(NC)- Docker logs for Prometheus (prod)\n"
	@printf "  $(GREEN)monitor-docker-logs-grafana$(NC)        - Docker logs for Grafana (local)\n"
	@printf "  $(GREEN)monitor-docker-logs-grafana-prod$(NC)   - Docker logs for Grafana (prod)\n"
	@printf "  $(GREEN)monitor-docker-logs-db$(NC)             - Docker logs for DB exporter (local)\n"
	@printf "  $(GREEN)monitor-docker-logs-db-prod$(NC)        - Docker logs for DB exporter (prod)\n"
	@printf "  $(GREEN)monitor-pull$(NC)                       - Pull latest monitoring images (local)\n"
	@printf "  $(GREEN)monitor-pull-prod$(NC)                  - Pull latest monitoring images (prod)\n\n"
	@printf "$(BOLD)$(BLUE)Status & Logs:$(NC)\n"
	@printf "  $(GREEN)monitor-status$(NC)                     - Show status of monitoring services\n"
	@printf "  $(GREEN)monitor-logs$(NC)                       - Show logs from all services (local)\n"
	@printf "  $(GREEN)monitor-logs-prod$(NC)                  - Show logs from all services (prod)\n\n"
	@printf "$(BOLD)$(BLUE)Testing:$(NC)\n"
	@printf "  $(GREEN)monitor-test$(NC)                       - Run full test suite (local only)\n"
	@printf "  $(GREEN)monitor-targets$(NC)                    - Show Prometheus targets status\n"
	@printf "  $(GREEN)monitor-traffic$(NC)                    - Generate test traffic (local)\n"
	@printf "  $(GREEN)monitor-traffic-heavy$(NC)              - Generate heavy test traffic (local)\n"
	@printf "  $(GREEN)monitor-traffic-prod$(NC)               - Generate test traffic (production)\n"
	@printf "  $(GREEN)monitor-traffic-heavy-prod$(NC)         - Generate heavy test traffic (prod)\n\n"
	@printf "$(BOLD)$(BLUE)Access:$(NC)\n"
	@printf "  $(GREEN)monitor-grafana$(NC)                    - Open Grafana in browser\n"
	@printf "  $(GREEN)monitor-prometheus$(NC)                 - Open Prometheus in browser\n"
	@printf "  $(GREEN)monitor-metrics$(NC)                    - Show all metrics endpoints\n"
	@printf "  $(GREEN)monitor-caddy-metrics$(NC)              - Show Caddy metrics\n"
	@printf "  $(GREEN)monitor-api-metrics$(NC)                - Show API metrics\n"
	@printf "  $(GREEN)monitor-db-metrics$(NC)                 - Show PostgreSQL metrics (local)\n"
	@printf "  $(GREEN)monitor-db-metrics-prod$(NC)            - Show PostgreSQL metrics (prod)\n\n"
	@printf "$(BOLD)$(BLUE)Utilities:$(NC)\n"
	@printf "  $(GREEN)monitor-stats$(NC)                      - Show resource usage (local)\n"
	@printf "  $(GREEN)monitor-stats-prod$(NC)                 - Show resource usage (prod)\n"
	@printf "  $(GREEN)monitor-config$(NC)                     - Show Prometheus config (local)\n"
	@printf "  $(GREEN)monitor-config-prod$(NC)                - Show Prometheus config (prod)\n"
	@printf "  $(GREEN)monitor-backup$(NC)                     - Backup Prometheus data (local)\n"
	@printf "  $(GREEN)monitor-backup-prod$(NC)                - Backup Prometheus data (prod)\n"
	@printf "  $(GREEN)monitor-export-dashboards$(NC)          - Export Grafana dashboards to JSON\n"
	@printf "  $(GREEN)monitor-clean$(NC)                      - Clean all monitoring data (local)\n"
	@printf "  $(GREEN)monitor-clean-prod$(NC)                 - Clean all monitoring data (prod)\n\n"
	@printf "$(BOLD)Quick Start:$(NC)\n"
	@printf "  1. $(YELLOW)make monitor-up$(NC)           - Start the stack\n"
	@printf "  2. $(YELLOW)make monitor-test$(NC)         - Verify everything works\n"
	@printf "  3. $(YELLOW)make monitor-traffic$(NC)      - Generate some traffic\n"
	@printf "  4. $(YELLOW)make monitor-grafana$(NC)      - Open dashboards\n\n"
	@printf "$(BOLD)Docker Compose Examples:$(NC)\n"
	@printf "  $(YELLOW)docker compose --profile local up -d$(NC)           - Start local stack\n"
	@printf "  $(YELLOW)docker compose --profile prod up -d$(NC)            - Start prod stack\n"
	@printf "  $(YELLOW)docker ps --filter name=prometheus$(NC)             - List containers\n"
	@printf "  $(YELLOW)docker exec -it oullin_prometheus_local /bin/sh$(NC) - Shell access\n\n"
