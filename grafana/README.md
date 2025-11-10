# Grafana Monitoring Dashboard

This directory contains the Grafana configuration for monitoring the Oullin application stack.

## Access

Grafana is accessible at [http://localhost:3000](http://localhost:3000) (from the server)

**Default Credentials:**
- Username: `admin`
- Password: Set via `GRAFANA_ADMIN_PASSWORD` environment variable (defaults to `admin`)

**Security Note:** Change the default password on first login or set `GRAFANA_ADMIN_PASSWORD` in your `.env` file.

## Remote Access

To access Grafana from your local machine:

```bash
ssh -L 3000:localhost:3000 user@your-server.com
```

Then open [http://localhost:3000](http://localhost:3000) in your browser.

## Pre-configured Dashboards

Three dashboards are automatically provisioned:

### 1. Oullin - Overview
High-level view of all services:
- Caddy request rate
- PostgreSQL active connections
- HTTP requests by status code
- API memory usage and goroutines

### 2. PostgreSQL - Database Metrics
Detailed database monitoring:
- Active connections
- Database size
- Transaction rates
- Database operations (inserts, updates, deletes)
- Cache hit ratio
- Lock statistics

### 3. Caddy - Proxy Metrics
Reverse proxy performance:
- Total request rate
- Active connections
- Response time percentiles (p50, p95, p99)
- Requests by status code
- Traffic rate (request/response sizes)
- Connection states

## Data Source

Grafana is pre-configured with Prometheus as the default data source, automatically connecting to the Prometheus service at `http://prometheus:9090`.

## Customization

Dashboards can be edited through the Grafana UI. To persist changes:

1. Edit the dashboard in Grafana
2. Click "Dashboard settings" → "JSON Model"
3. Copy the JSON
4. Save to `./grafana/dashboards/your-dashboard.json`
5. Restart Grafana to load changes

## Directory Structure

```text
grafana/
├── README.md
├── dashboards/               # Dashboard JSON files
│   ├── oullin-overview-oullin-overview.json
│   ├── oullin-postgresql-postgresql-database-metrics.json
│   └── oullin-caddy-caddy-proxy-metrics.json
└── provisioning/
    ├── datasources/          # Data source configuration
    │   └── prometheus.yml
    └── dashboards/           # Dashboard provisioning config
        └── default.yml
```

## Useful Queries

### API Metrics
```promql
# Request rate
rate(promhttp_metric_handler_requests_total[5m])

# Memory usage
go_memstats_alloc_bytes{job="api"}

# Goroutines
go_goroutines{job="api"}
```

### Database Metrics
```promql
# Connection count
pg_stat_database_numbackends

# Transaction rate
rate(pg_stat_database_xact_commit[5m])

# Database size
pg_database_size_bytes

# Cache hit ratio
rate(pg_stat_database_blks_hit[5m]) / (rate(pg_stat_database_blks_hit[5m]) + rate(pg_stat_database_blks_read[5m]))
```

### Caddy Metrics
```promql
# Request rate
rate(caddy_http_request_count_total[5m])

# Response time (95th percentile)
histogram_quantile(0.95, rate(caddy_http_request_duration_seconds_bucket[5m]))

# Response traffic rate
rate(caddy_http_response_size_bytes_sum[5m])

# Error rate
rate(caddy_http_request_errors_total[5m])
```

## Troubleshooting

If dashboards don't load:
1. Check Grafana logs: `docker logs oullin_grafana`
2. Verify Prometheus connection: Settings → Data Sources → Prometheus → "Save & Test"
3. Ensure Prometheus is running: `docker ps | grep prometheus`

If no data appears:
1. Verify Prometheus is scraping targets: http://localhost:9090/targets
2. Check that services are exposing metrics
3. Wait a few minutes for initial data collection
