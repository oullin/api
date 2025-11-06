# -------------------------------------------------------------------------------------------------------------------- #
# Monitoring Stack Targets
# -------------------------------------------------------------------------------------------------------------------- #

# -------------------------------------------------------------------------------------------------------------------- #
# Start/Stop Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Start monitoring stack (local development)
monitor\:up:
	@printf "$(BOLD)$(CYAN)Starting monitoring stack (local)...$(NC)\n"
	@docker compose --profile local up -d prometheus_local grafana_local postgres_exporter_local
	@sleep 3
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack started$(NC)\n"
	@printf "\n$(BOLD)Access points:$(NC)\n"
	@printf "  $(GREEN)Grafana:$(NC)     http://localhost:3000\n"
	@printf "  $(GREEN)Prometheus:$(NC)  http://localhost:9090\n"
	@printf "  $(GREEN)Caddy Admin:$(NC) http://localhost:2019\n\n"

## Start monitoring stack (production)
monitor\:up\:prod:
	@printf "$(BOLD)$(CYAN)Starting monitoring stack (production)...$(NC)\n"
	@docker compose --profile prod up -d prometheus grafana postgres_exporter
	@sleep 3
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack started$(NC)\n"
	@printf "\n$(BOLD)Access points (from server):$(NC)\n"
	@printf "  $(GREEN)Grafana:$(NC)     http://localhost:3000\n"
	@printf "  $(GREEN)Prometheus:$(NC)  http://localhost:9090\n"
	@printf "  $(GREEN)Caddy Admin:$(NC) http://localhost:2019\n\n"

## Stop monitoring stack (local)
monitor\:down:
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack (local)...$(NC)\n"
	@docker compose --profile local stop prometheus_local grafana_local postgres_exporter_local
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack stopped$(NC)\n\n"

## Stop monitoring stack (production)
monitor\:down\:prod:
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack (production)...$(NC)\n"
	@docker compose --profile prod stop prometheus grafana postgres_exporter
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack stopped$(NC)\n\n"

## Restart monitoring stack (local)
monitor\:restart:
	@printf "$(BOLD)$(CYAN)Restarting monitoring stack...$(NC)\n"
	@docker compose --profile local restart prometheus_local grafana_local postgres_exporter_local
	@printf "$(BOLD)$(GREEN)✓ Monitoring stack restarted$(NC)\n\n"

# -------------------------------------------------------------------------------------------------------------------- #
# Status & Information Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Show status of monitoring services
monitor\:status:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Status$(NC)\n\n"
	@docker ps --filter "name=prometheus" --filter "name=grafana" --filter "name=exporter" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@printf "\n"

## Show logs from all monitoring services
monitor\:logs:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Logs$(NC)\n\n"
	@docker compose logs -f prometheus_local grafana_local postgres_exporter_local

## Show Prometheus logs
monitor\:logs\:prometheus:
	@docker logs -f oullin_prometheus_local

## Show Grafana logs
monitor\:logs\:grafana:
	@docker logs -f oullin_grafana_local

## Show PostgreSQL exporter logs
monitor\:logs\:db:
	@docker logs -f oullin_postgres_exporter_local

# -------------------------------------------------------------------------------------------------------------------- #
# Testing & Verification Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Run full monitoring stack test suite
monitor\:test:
	@printf "$(BOLD)$(CYAN)Running monitoring stack tests...$(NC)\n\n"
	@printf "$(BOLD)1. Checking services are running...$(NC)\n"
	@docker ps --filter "name=prometheus_local" --filter "name=grafana_local" --filter "name=postgres_exporter_local" --format "  ✓ {{.Names}}: {{.Status}}" || echo "  $(RED)✗ Services not running$(NC)"
	@printf "\n$(BOLD)2. Testing Prometheus targets...$(NC)\n"
	@curl -s http://localhost:9090/api/v1/targets | grep -q '"health":"up"' && echo "  $(GREEN)✓ Prometheus targets are UP$(NC)" || echo "  $(RED)✗ Some targets are DOWN$(NC)"
	@printf "\n$(BOLD)3. Testing Caddy metrics endpoint...$(NC)\n"
	@curl -s http://localhost:2019/metrics | grep -q "caddy_http_requests_total" && echo "  $(GREEN)✓ Caddy metrics accessible$(NC)" || echo "  $(RED)✗ Caddy metrics unavailable$(NC)"
	@printf "\n$(BOLD)4. Testing API metrics endpoint...$(NC)\n"
	@curl -s http://localhost:8080/metrics | grep -q "go_goroutines" && echo "  $(GREEN)✓ API metrics accessible$(NC)" || echo "  $(RED)✗ API metrics unavailable$(NC)"
	@printf "\n$(BOLD)5. Testing Grafana...$(NC)\n"
	@curl -s http://localhost:3000/api/health | grep -q "ok" && echo "  $(GREEN)✓ Grafana is healthy$(NC)" || echo "  $(RED)✗ Grafana is unhealthy$(NC)"
	@printf "\n$(BOLD)$(GREEN)Test suite completed!$(NC)\n\n"

## Verify Prometheus targets status
monitor\:targets:
	@printf "$(BOLD)$(CYAN)Prometheus Targets Status$(NC)\n\n"
	@curl -s http://localhost:9090/api/v1/targets | jq -r '.data.activeTargets[] | "[\(.health | ascii_upcase)] \(.labels.job) - \(.scrapeUrl)"' || echo "$(RED)Failed to fetch targets. Is Prometheus running?$(NC)"
	@printf "\n"

## Check Prometheus configuration
monitor\:config:
	@printf "$(BOLD)$(CYAN)Prometheus Configuration$(NC)\n\n"
	@docker exec oullin_prometheus_local cat /etc/prometheus/prometheus.yml

# -------------------------------------------------------------------------------------------------------------------- #
# Metrics Access Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Open Grafana in browser
monitor\:grafana:
	@printf "$(BOLD)$(CYAN)Opening Grafana...$(NC)\n"
	@printf "URL: $(GREEN)http://localhost:3000$(NC)\n"
	@printf "Credentials: admin / (set via GRAFANA_ADMIN_PASSWORD)\n\n"
	@which xdg-open > /dev/null && xdg-open http://localhost:3000 || which open > /dev/null && open http://localhost:3000 || echo "Please open http://localhost:3000 in your browser"

## Open Prometheus in browser
monitor\:prometheus:
	@printf "$(BOLD)$(CYAN)Opening Prometheus...$(NC)\n"
	@printf "URL: $(GREEN)http://localhost:9090$(NC)\n\n"
	@which xdg-open > /dev/null && xdg-open http://localhost:9090 || which open > /dev/null && open http://localhost:9090 || echo "Please open http://localhost:9090 in your browser"

## Show Caddy metrics
monitor\:caddy-metrics:
	@printf "$(BOLD)$(CYAN)Caddy Metrics$(NC)\n\n"
	@curl -s http://localhost:2019/metrics | grep "^caddy_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n"
	@printf "Full metrics: $(GREEN)http://localhost:2019/metrics$(NC)\n\n"

## Show API metrics
monitor\:api-metrics:
	@printf "$(BOLD)$(CYAN)API Metrics$(NC)\n\n"
	@curl -s http://localhost:8080/metrics | grep "^go_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n"
	@printf "Full metrics: $(GREEN)http://localhost:8080/metrics$(NC)\n\n"

## Show PostgreSQL metrics
monitor\:db-metrics:
	@printf "$(BOLD)$(CYAN)PostgreSQL Metrics$(NC)\n\n"
	@docker exec oullin_prometheus_local curl -s http://postgres_exporter_local:9187/metrics | grep "^pg_" | head -20
	@printf "\n$(YELLOW)... (showing first 20 metrics)$(NC)\n\n"

## Show all metrics endpoints
monitor\:metrics:
	@printf "$(BOLD)$(CYAN)Available Metrics Endpoints$(NC)\n\n"
	@printf "  $(GREEN)Caddy:$(NC)      http://localhost:2019/metrics\n"
	@printf "  $(GREEN)API:$(NC)        http://localhost:8080/metrics\n"
	@printf "  $(GREEN)PostgreSQL:$(NC) http://postgres_exporter_local:9187/metrics (internal)\n"
	@printf "  $(GREEN)Prometheus:$(NC) http://localhost:9090/metrics\n\n"

# -------------------------------------------------------------------------------------------------------------------- #
# Traffic Generation & Testing
# -------------------------------------------------------------------------------------------------------------------- #

## Generate test traffic to populate metrics
monitor\:traffic:
	@printf "$(BOLD)$(CYAN)Generating test traffic...$(NC)\n"
	@printf "Making 100 requests to /ping endpoint...\n"
	@for i in {1..100}; do \
		curl -s http://localhost:8080/ping > /dev/null && printf "." || printf "$(RED)✗$(NC)"; \
		sleep 0.1; \
	done
	@printf "\n$(BOLD)$(GREEN)✓ Test traffic generated$(NC)\n"
	@printf "\nCheck dashboards at: $(GREEN)http://localhost:3000$(NC)\n\n"

## Generate heavy test traffic
monitor\:traffic\:heavy:
	@printf "$(BOLD)$(CYAN)Generating heavy test traffic...$(NC)\n"
	@printf "Making 500 requests with 5 concurrent connections...\n"
	@for i in {1..100}; do \
		(for j in {1..5}; do curl -s http://localhost:8080/ping > /dev/null & done; wait); \
		printf "."; \
		sleep 0.05; \
	done
	@printf "\n$(BOLD)$(GREEN)✓ Heavy test traffic generated$(NC)\n\n"

# -------------------------------------------------------------------------------------------------------------------- #
# Utility Commands
# -------------------------------------------------------------------------------------------------------------------- #

## Clean monitoring data (removes all metrics/dashboard data)
monitor\:clean:
	@printf "$(BOLD)$(RED)WARNING: This will delete all monitoring data!$(NC)\n"
	@printf "Press Ctrl+C to cancel, or Enter to continue..."
	@read
	@printf "$(BOLD)$(CYAN)Stopping monitoring stack...$(NC)\n"
	@docker compose --profile local down prometheus_local grafana_local
	@printf "$(BOLD)$(CYAN)Removing volumes...$(NC)\n"
	@docker volume rm -f prometheus_data grafana_data || true
	@printf "$(BOLD)$(GREEN)✓ Monitoring data cleaned$(NC)\n\n"

## Show monitoring stack resource usage
monitor\:stats:
	@printf "$(BOLD)$(CYAN)Monitoring Stack Resource Usage$(NC)\n\n"
	@docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" \
		oullin_prometheus_local oullin_grafana_local oullin_postgres_exporter_local 2>/dev/null || \
		echo "$(RED)No monitoring containers running$(NC)"
	@printf "\n"

## Backup Prometheus data
monitor\:backup:
	@printf "$(BOLD)$(CYAN)Backing up Prometheus data...$(NC)\n"
	@mkdir -p ./backups
	@docker run --rm -v prometheus_data:/data -v $(PWD)/backups:/backup alpine \
		tar czf /backup/prometheus-backup-$$(date +%Y%m%d-%H%M%S).tar.gz /data
	@printf "$(BOLD)$(GREEN)✓ Backup created in ./backups/$(NC)\n\n"

## Show monitoring help
monitor\:help:
	@printf "\n$(BOLD)$(CYAN)Monitoring Stack Commands$(NC)\n\n"
	@printf "$(BOLD)$(BLUE)Start/Stop:$(NC)\n"
	@printf "  $(GREEN)monitor:up$(NC)              - Start monitoring stack (local)\n"
	@printf "  $(GREEN)monitor:up:prod$(NC)         - Start monitoring stack (production)\n"
	@printf "  $(GREEN)monitor:down$(NC)            - Stop monitoring stack (local)\n"
	@printf "  $(GREEN)monitor:down:prod$(NC)       - Stop monitoring stack (production)\n"
	@printf "  $(GREEN)monitor:restart$(NC)         - Restart monitoring stack\n\n"
	@printf "$(BOLD)$(BLUE)Status & Logs:$(NC)\n"
	@printf "  $(GREEN)monitor:status$(NC)          - Show status of monitoring services\n"
	@printf "  $(GREEN)monitor:logs$(NC)            - Show logs from all services\n"
	@printf "  $(GREEN)monitor:logs:prometheus$(NC) - Show Prometheus logs\n"
	@printf "  $(GREEN)monitor:logs:grafana$(NC)    - Show Grafana logs\n"
	@printf "  $(GREEN)monitor:logs:db$(NC)         - Show PostgreSQL exporter logs\n\n"
	@printf "$(BOLD)$(BLUE)Testing:$(NC)\n"
	@printf "  $(GREEN)monitor:test$(NC)            - Run full test suite\n"
	@printf "  $(GREEN)monitor:targets$(NC)         - Show Prometheus targets status\n"
	@printf "  $(GREEN)monitor:traffic$(NC)         - Generate test traffic\n"
	@printf "  $(GREEN)monitor:traffic:heavy$(NC)   - Generate heavy test traffic\n\n"
	@printf "$(BOLD)$(BLUE)Access:$(NC)\n"
	@printf "  $(GREEN)monitor:grafana$(NC)         - Open Grafana in browser\n"
	@printf "  $(GREEN)monitor:prometheus$(NC)      - Open Prometheus in browser\n"
	@printf "  $(GREEN)monitor:metrics$(NC)         - Show all metrics endpoints\n"
	@printf "  $(GREEN)monitor:caddy-metrics$(NC)   - Show Caddy metrics\n"
	@printf "  $(GREEN)monitor:api-metrics$(NC)     - Show API metrics\n"
	@printf "  $(GREEN)monitor:db-metrics$(NC)      - Show PostgreSQL metrics\n\n"
	@printf "$(BOLD)$(BLUE)Utilities:$(NC)\n"
	@printf "  $(GREEN)monitor:stats$(NC)           - Show resource usage\n"
	@printf "  $(GREEN)monitor:config$(NC)          - Show Prometheus config\n"
	@printf "  $(GREEN)monitor:backup$(NC)          - Backup Prometheus data\n"
	@printf "  $(GREEN)monitor:clean$(NC)           - Clean all monitoring data\n\n"
	@printf "$(BOLD)Quick Start:$(NC)\n"
	@printf "  1. $(YELLOW)make monitor:up$(NC)      - Start the stack\n"
	@printf "  2. $(YELLOW)make monitor:test$(NC)    - Verify everything works\n"
	@printf "  3. $(YELLOW)make monitor:traffic$(NC) - Generate some traffic\n"
	@printf "  4. $(YELLOW)make monitor:grafana$(NC) - Open dashboards\n\n"
