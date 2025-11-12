# Monitoring Stack Documentation

Complete guide for managing and monitoring the Oullin application stack with Prometheus, Grafana, and related tools.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Security Model](#security-model)
4. [Grafana Dashboards](#grafana-dashboards)
5. [Creating Custom Dashboards](#creating-custom-dashboards)
6. [Prometheus Queries](#prometheus-queries)
7. [Troubleshooting](#troubleshooting)
8. [Maintenance & Backup](#maintenance--backup)
9. [Resources](#resources)

**For VPS deployment instructions, see [VPS_DEPLOYMENT.md](./VPS_DEPLOYMENT.md)**

---

## Overview

### Stack Components

- **Prometheus**: Metrics collection and time-series storage
- **Grafana**: Visualization dashboards and alerting
- **postgres_exporter**: PostgreSQL database metrics
- **Caddy Admin API**: Reverse proxy metrics

### Pre-configured Dashboards

Three dashboards are automatically provisioned:

1. **Oullin - Overview** (`grafana/dashboards/oullin-overview-oullin-overview.json`)
   - Caddy request rate
   - PostgreSQL active connections
   - HTTP requests by status code
   - API memory usage and goroutines

2. **PostgreSQL - Database Metrics** (`grafana/dashboards/oullin-postgresql-postgresql-database-metrics.json`)
   - Active connections
   - Database size
   - Transaction rates
   - Cache hit ratio
   - Lock statistics

3. **Caddy - Proxy Metrics** (`grafana/dashboards/oullin-caddy-caddy-proxy-metrics.json`)
   - Total request rate
   - Response time percentiles
   - Requests by status code
   - Traffic rate
   - Request errors

### Directory Structure

```text
infra/metrics/
├── README.md                    # This file
├── grafana/
│   ├── dashboards/              # Dashboard JSON files
│   ├── provisioning/
│   │   ├── dashboards/          # Dashboard provisioning config
│   │   └── datasources/         # Data source configuration
│   └── scripts/
│       └── export-dashboards.sh
└── prometheus/
    ├── provisioning/
    │   ├── prometheus.yml       # Production Prometheus config
    │   └── prometheus.local.yml # Local Prometheus config
    └── scripts/
        └── postgres-exporter-entrypoint.sh
```

### Configuration Consistency

The monitoring stack is designed to maintain configuration consistency across local and production environments while respecting environment-specific differences.

#### Shared Configuration Elements

The following configurations are **identical** across both environments:

1. **Grafana Settings:**
   - Same Grafana version (`grafana/grafana:11.4.0`)
   - Identical security settings (admin user, sign-up disabled, anonymous disabled)
   - Same dashboard and datasource provisioning structure
   - Same volume mount paths

2. **Prometheus Core Settings:**
   - Same Prometheus version (`prom/prometheus:v3.0.1`)
   - Identical scrape interval (15s) and evaluation interval (15s)
   - Same job structure (caddy, postgresql, api, prometheus) with per-environment targets
   - Same metrics endpoints and paths

3. **Postgres Exporter:**
   - Same exporter version (`prometheuscommunity/postgres-exporter:v0.15.0`)
   - Identical port exposure (9187)
   - Same entrypoint script and secrets handling

#### Environment-Specific Variables

These settings **differ intentionally** based on environment:

| Configuration | Local | Production | Reason |
|--------------|-------|------------|--------|
| **Container Names** | `oullin_*_local` | `oullin_*` | Distinguish environments |
| **Prometheus URL** | `oullin_prometheus_local:9090` | `oullin_prometheus:9090` | Network addressing |
| **Grafana Port** | `3000:3000` | `127.0.0.1:3000:3000` | Security (prod localhost-only) |
| **Prometheus Port** | `9090:9090` | `127.0.0.1:9090:9090` | Security (prod localhost-only) |
| **Data Retention** | 7 days | 30 days | Storage/cost optimization |
| **Caddy Target** | `caddy_local:9180` | `caddy_prod:9180` | Service dependencies |
| **PostgreSQL Exporter Target** | `oullin_postgres_exporter_local:9187` | `oullin_postgres_exporter:9187` | Service dependencies |
| **External Labels** | `monitor: 'oullin-local'`<br>`environment: 'local'` | `monitor: 'oullin-prod'`<br>`environment: 'production'` | Metric identification |
| **Admin API** | `127.0.0.1:2019:2019` | Not exposed | Debugging access |

#### Environment Variable Usage

The configuration uses environment variables to maintain consistency while adapting to each environment:

**Grafana Datasource** (`grafana/provisioning/datasources/prometheus.yml`):
```yaml
url: ${GF_DATASOURCE_PROMETHEUS_URL}
```

Set via Docker Compose:
- **Local:** `GF_DATASOURCE_PROMETHEUS_URL=http://oullin_prometheus_local:9090`
- **Production:** `GF_DATASOURCE_PROMETHEUS_URL=http://oullin_prometheus:9090`

**Required Environment Variables:**
- `GRAFANA_ADMIN_PASSWORD` - **Required**, no default (set in `.env`)
- `GF_DATASOURCE_PROMETHEUS_URL` - Set automatically by Docker Compose profile

#### Configuration Files by Environment

**Local Environment:**
- Prometheus: `prometheus/provisioning/prometheus.local.yml`
- Profile: `--profile local`
- Services: `prometheus_local`, `grafana_local`, `caddy_local`, `postgres_exporter_local`

**Production Environment:**
- Prometheus: `prometheus/provisioning/prometheus.yml`
- Profile: `--profile prod`
- Services: `prometheus`, `grafana`, `caddy_prod`, `postgres_exporter`

**Shared Across All Environments:**
- Grafana datasources: `grafana/provisioning/datasources/prometheus.yml`
- Grafana dashboards: `grafana/provisioning/dashboards/default.yml`
- Dashboard JSONs: `grafana/dashboards/*.json`
- Postgres exporter script: `prometheus/scripts/postgres-exporter-entrypoint.sh`

---

## Quick Start

### Local Development

**Prerequisites:**
- Docker and Docker Compose installed
- `.env` file in the repository root with `GRAFANA_ADMIN_PASSWORD` set (required - no default)
  - Use `make env:init` to copy `.env.example` if you need a starting point
  - If `.env` already exists, edit it in place instead of appending duplicates
- Database secrets in `database/infra/secrets/`

**Setup:**

```bash
# 1. Set Grafana admin password in .env file
echo "GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 32)" >> .env
# (Add or update the key manually if the file already defines it.)

# 2. Start the local monitoring stack
make monitor-up
# Or: docker compose --profile local up -d

# 3. Access services
# Grafana:    http://localhost:3000 (admin / your-password)
# Prometheus: http://localhost:9090
# Caddy Admin: http://localhost:2019
```

**Verification:**

```bash
# Check all services are running
docker ps

# Verify Prometheus targets are UP
make monitor-targets
# Or: curl http://localhost:9090/api/v1/targets

# Generate test traffic
make monitor-traffic

# View dashboards
make monitor-grafana
```

---

## Security Model

### Critical Security Requirements

⚠️ **IMPORTANT**: The monitoring stack includes several security considerations:

1. **Grafana Admin Password**
   - No default password allowed
   - Must set `GRAFANA_ADMIN_PASSWORD` in `.env`
   - Docker Compose will fail if not set
   - Generate strong password: `openssl rand -base64 32`

2. **Caddy Admin API**
   - Exposes powerful administrative endpoints (`/load`, `/config`, `/stop`)
   - **NO authentication** by default
   - Production: Only accessible within Docker network; restrict further via firewalls/security groups when possible
   - If you must expose it, configure Caddy's admin access controls (`admin.identity`, `admin.authorize`, or reverse-proxy ACLs) to require authentication
   - Never expose to public internet

3. **Service Exposure**
   - Production: Services bound to `127.0.0.1` only
   - Access via SSH tunneling from remote
   - No direct internet exposure

### Production Security Configuration

**Docker Compose Production Services:**

```yaml
grafana:
  ports:
    - "127.0.0.1:3000:3000"  # Localhost only

prometheus:
  ports:
    - "127.0.0.1:9090:9090"  # Localhost only

caddy_prod:
  expose:
    - "2019"  # Internal network only - NOT exposed to host
```

**Remote Access:**

```bash
# SSH tunnel for Grafana and Prometheus
ssh -L 3000:localhost:3000 -L 9090:localhost:9090 user@your-server

# Access Caddy admin API (debugging only)
docker exec -it oullin_proxy_prod curl http://localhost:2019/metrics
```

### Security Checklist

- ✅ `GRAFANA_ADMIN_PASSWORD` set with strong password
- ✅ Firewall configured (UFW)
- ✅ Only necessary ports exposed (22, 80, 443)
- ✅ Monitoring services NOT exposed to internet
- ✅ Docker secrets for sensitive data
- ✅ Regular backups scheduled
- ✅ Log rotation configured
- ✅ SSH key-based authentication

---

## Grafana Dashboards

### Accessing Dashboards

**Local:** <http://localhost:3000>
**Production:** SSH tunnel then <http://localhost:3000>

### Dashboard Files

All dashboards are in `infra/metrics/grafana/dashboards/`:
- `oullin-overview-oullin-overview.json`
- `oullin-postgresql-postgresql-database-metrics.json`
- `oullin-caddy-caddy-proxy-metrics.json`

### Exporting Dashboards

Use the built-in export script:

```bash
make monitor-export-dashboards
```

This will:
1. List all dashboards in Grafana
2. Let you select which to export
3. Save to `infra/metrics/grafana/dashboards/`
4. Format properly for provisioning

### Manual Export

1. Open your dashboard in Grafana
2. Click **"Share"** → **"Export"** tab
3. Click **"Save to file"** or **"View JSON"**
4. Save to `infra/metrics/grafana/dashboards/`
5. Restart Grafana: `make monitor-restart`

### Updating Dashboards Safely

To keep dashboard changes reproducible and under version control:

1. **Start monitoring stack**: `make monitor-up`
2. **Make changes in Grafana UI**: Navigate to <http://localhost:3000> and edit dashboards
3. **Export your changes**: Run `./infra/metrics/grafana/scripts/export-dashboards.sh`
   - Select specific dashboard or `all` to export all dashboards
   - Exports are saved to `infra/metrics/grafana/dashboards/`
4. **Review the diff**: `git diff infra/metrics/grafana/dashboards/`
5. **Commit changes**: Add and commit the exported JSON files
6. **Verify**: `make monitor-restart` to ensure dashboards reload correctly

**Warning:** Always export after making UI changes—manual edits to JSON files can work but are error-prone.

---

## Creating Custom Dashboards

### Method 1: Create in UI (Recommended)

**Step 1:** Start Grafana

```bash
make monitor-up
make monitor-grafana  # Opens http://localhost:3000
```

**Step 2:** Create dashboard

1. Click **"+"** → **"Dashboard"** → **"Add visualization"**
2. Select **"Prometheus"** as data source
3. Write PromQL query
4. Choose visualization type (Time series, Stat, Gauge, Table)
5. Configure panel (title, description, units, thresholds)
6. Add more panels as needed
7. Save dashboard

**Step 3:** Export

```bash
make monitor-export-dashboards
```

### Method 2: Use Community Dashboards

Grafana has thousands of pre-built dashboards at <https://grafana.com/grafana/dashboards/>

**Popular for our stack:**
- [9628](https://grafana.com/grafana/dashboards/9628) - PostgreSQL Database
- [455](https://grafana.com/grafana/dashboards/455) - PostgreSQL Stats
- [10826](https://grafana.com/grafana/dashboards/10826) - Go Metrics
- [6671](https://grafana.com/grafana/dashboards/6671) - Go Processes

**Import via UI:**
1. Click **"+"** → **"Import"**
2. Enter dashboard ID
3. Select **"Prometheus"** as data source
4. Click **"Import"**

### Dashboard Best Practices

**Organization:**
- One dashboard per service
- Overview dashboard for high-level metrics
- Detail dashboards for deep dives
- Use tags for categorization

**Panel Design:**
- Clear titles
- Descriptions for complex metrics
- Consistent colors
- Appropriate units (bytes, %, req/s)
- Thresholds for warnings/errors

**Query Performance:**
- Avoid high-cardinality labels
- Use recording rules for expensive queries
- Limit time range
- Use `rate()` instead of raw counters

---

## Prometheus Queries

### API Metrics

```promql
# Request rate
rate(promhttp_metric_handler_requests_total[5m])

# Memory usage
go_memstats_alloc_bytes{job="api"}

# Goroutines (check for leaks)
go_goroutines{job="api"}

# GC duration
rate(go_gc_duration_seconds_sum[5m])

# Heap allocations
rate(go_memstats_alloc_bytes_total[5m])
```

### PostgreSQL Metrics

```promql
# Active connections
pg_stat_database_numbackends

# Database size
pg_database_size_bytes

# Transaction rate
rate(pg_stat_database_xact_commit[5m])

# Cache hit ratio (should be >90%)
rate(pg_stat_database_blks_hit[5m]) /
(rate(pg_stat_database_blks_hit[5m]) + rate(pg_stat_database_blks_read[5m]))

# Rows inserted/updated/deleted
rate(pg_stat_database_tup_inserted[5m])
rate(pg_stat_database_tup_updated[5m])
rate(pg_stat_database_tup_deleted[5m])
```

### Caddy Metrics

```promql
# Request rate by status
sum by(code) (rate(caddy_http_requests_total[5m]))

# Response time percentiles
histogram_quantile(0.95, rate(caddy_http_request_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(caddy_http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(caddy_http_request_errors_total[5m]))

# Response traffic rate
rate(caddy_http_response_size_bytes_sum[5m])
```

---

## Troubleshooting

### Dashboards Don't Load

```bash
# Check JSON syntax
jq . < infra/metrics/grafana/dashboards/my-dashboard.json

# Check Grafana logs
docker logs oullin_grafana_local  # Local
docker logs oullin_grafana        # Production

# Or view all monitoring logs
make monitor-logs      # Local
make monitor-logs-prod # Production

# Verify Prometheus connection
# Grafana UI → Settings → Data Sources → Prometheus → "Save & Test"

# Ensure Prometheus is running
docker ps | grep prometheus
```

### No Data in Panels

```bash
# Verify Prometheus is scraping targets
make monitor-targets
# Or: curl http://localhost:9090/api/v1/targets

# Test query in Prometheus
# Open http://localhost:9090

# Wait a few minutes for initial data collection
```

### Prometheus Not Scraping

```bash
# Check network connectivity
docker exec -it oullin_prometheus_local ping caddy_local

# Verify service exposes metrics
docker exec -it oullin_prometheus_local curl http://caddy_local:2019/metrics

# Check Prometheus config
docker exec -it oullin_prometheus_local cat /etc/prometheus/prometheus.yml
```

### Targets Show as DOWN

```bash
# Check container networking
docker network ls
docker network inspect caddy_net

# Check container names match Prometheus config
docker ps

# Restart services
make monitor-restart
# Or: docker compose --profile local restart
```

### High Memory Usage

```bash
# Monitor memory
docker stats

# If Prometheus using too much memory:
# - Reduce retention time
# - Decrease scrape frequency
# - Add metric filters
```

### Data Not Persisting

```bash
# Ensure volumes are configured
docker volume ls
docker volume inspect prometheus_data_local   # Local
docker volume inspect prometheus_data_prod    # Production
docker volume inspect grafana_data_local      # Local
docker volume inspect grafana_data_prod       # Production
```

---

## Maintenance & Backup

### Backing Up Data

**Automated backup** (recommended):

```bash
# Runs daily via cron, keeps last 5 backups
make monitor-backup       # Local environment
make monitor-backup-prod  # Production environment
```

Backups saved to:
- **Local**: `storage/monitoring/backups/prometheus-backup-YYYYMMDD-HHMMSS.tar.gz`
- **Production**: `storage/monitoring/backups/prometheus-prod-backup-YYYYMMDD-HHMMSS.tar.gz`

**Manual backup:**

```bash
# Backup Prometheus data
docker run --rm -v prometheus_data_local:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/prometheus-backup-$(date +%Y%m%d-%H%M%S).tar.gz /data
# (Use prometheus_data_prod on production hosts)

# Backup Grafana data
docker run --rm -v grafana_data_local:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/grafana-backup-$(date +%Y%m%d-%H%M%S).tar.gz /data
# (Use grafana_data_prod on production hosts)
```

### Restoring from Backup

```bash
# Stop services
make monitor-down

# Restore Prometheus data
# WARNING: This will DELETE all existing Prometheus data. Validate backups and consider restoring in a test environment first.
docker run --rm -v prometheus_data_local:/data -v $(pwd)/backups:/backup alpine \
  sh -c "rm -rf /data/* && tar xzf /backup/prometheus-backup-YYYYMMDD-HHMMSS.tar.gz -C /"
# (Use prometheus_data_prod on production hosts)

# Restore Grafana data
# WARNING: This will DELETE all existing Grafana data. Keep a secondary backup if unsure.
docker run --rm -v grafana_data_local:/data -v $(pwd)/backups:/backup alpine \
  sh -c "rm -rf /data/* && tar xzf /backup/grafana-backup-YYYYMMDD-HHMMSS.tar.gz -C /"
# (Use grafana_data_prod on production hosts)

# Restart services
make monitor-up
```

### Updating the Stack

**Local environment:**
```bash
# Pull latest images
docker compose pull

# Restart with new images
make monitor-restart
# Or: docker compose --profile local up -d
```

**Production environment:**
```bash
# Pull latest images
docker compose pull

# Restart with new images
make monitor-restart-prod
# Or: docker compose --profile prod up -d
```

### Monitoring Resource Usage

```bash
# CPU and Memory usage
docker stats

# Disk usage by container
docker system df -v

# Container logs size
sudo du -sh /var/lib/docker/containers/*/*-json.log
```

### Cleaning Up Old Data

Prometheus automatically handles retention based on `--storage.tsdb.retention.time` (30d prod, 7d local).

Manual cleanup:

```bash
# Stop Prometheus
docker compose stop prometheus_local

# Clean data
docker run --rm -v prometheus_data_local:/data alpine rm -rf /data/*
# (Use prometheus_data_prod on production hosts)

# Restart
docker compose --profile local up -d prometheus_local
```

---

## Resources

### Official Documentation

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [Caddy Metrics](https://caddyserver.com/docs/metrics)
- [PostgreSQL Exporter](https://github.com/prometheus-community/postgres_exporter)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafonnet Library](https://github.com/grafana/grafonnet-lib)

### Quick Reference Commands

```bash
# Start monitoring stack
make monitor-up              # Local
make monitor-up-prod         # Production

# Access services
make monitor-grafana         # Open Grafana
make monitor-prometheus      # Open Prometheus

# Check status
make monitor-status          # Service health
make monitor-targets         # Prometheus targets

# Generate traffic
make monitor-traffic         # Local
make monitor-traffic-prod    # Production

# View logs
make monitor-logs            # All services (local)
make monitor-logs-prod       # All services (production)

# Individual container logs
docker logs oullin_grafana_local     # Grafana (local)
docker logs oullin_prometheus_local  # Prometheus (local)
docker logs oullin_grafana           # Grafana (production)
docker logs oullin_prometheus        # Prometheus (production)

# Maintenance
make monitor-backup          # Backup Prometheus data
make monitor-restart         # Restart services (local)
make monitor-restart-prod    # Restart services (production)
make monitor-export-dashboards

# Cleanup
make monitor-down            # Stop services (local)
make monitor-down-prod       # Stop services (production)
make monitor-clean           # Clean up data (local)
make monitor-clean-prod      # Clean up data (production)
```

### Production Deployment

For complete VPS deployment instructions including firewall setup, SSL configuration, and production best practices, see [VPS_DEPLOYMENT.md](./VPS_DEPLOYMENT.md).

---

## Next Steps

1. **Set up Alerting**: Configure Prometheus Alertmanager for critical metrics
2. **Add Custom Metrics**: Instrument your API with custom business metrics
3. **Create Custom Dashboards**: Build dashboards specific to your use case
4. **Configure Recording Rules**: Pre-compute expensive queries
5. **Implement SLOs**: Define and monitor Service Level Objectives
6. **Export and Share**: Share dashboard configurations with your team

---

For questions or issues, please check the [Troubleshooting](#troubleshooting) section or refer to the official documentation links above.
