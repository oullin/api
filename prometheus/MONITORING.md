# Monitoring Stack Setup & Testing Guide

This document provides instructions for running and testing the monitoring stack both locally and in production.

## Stack Overview

The monitoring stack consists of:
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization dashboards
- **postgres_exporter**: PostgreSQL metrics exporter
- **Caddy Admin API**: Proxy metrics endpoint

## Security Model

### Caddy Admin API Security

**CRITICAL**: The Caddy admin API exposes powerful administrative endpoints (`/load`, `/config`, `/stop`) with **no authentication**. Improper exposure could allow unauthorized control of your reverse proxy.

#### Production Configuration

In production, the admin API is configured for **internal network access only**:

1. **Inside Container**: Bound to `0.0.0.0:2019` in `Caddyfile.prod`
   - Allows Prometheus to scrape metrics via Docker DNS (`caddy_prod:2019`)
   - Other containers in `caddy_net` can access it (acceptable risk within trusted network)

2. **Host Exposure**: Port 2019 is **NOT** exposed to the host in `docker-compose.yml`
   - No `ports` mapping for 2019 in production
   - The admin API is only accessible within the Docker network
   - Prevents unauthorized access from the host or public internet

#### Local Configuration

For local development, limited host access is provided for debugging:

- Port 2019 is exposed to `127.0.0.1` only
- Allows local debugging: `curl http://localhost:2019/metrics`
- Not exposed to external network interfaces

#### Security Best Practices

✅ **DO**:
- Keep admin API within Docker networks only in production
- Use SSH tunneling for remote access: `ssh -L 2019:localhost:2019 user@server`
- Monitor admin API access logs

❌ **DON'T**:
- Never expose admin API to `0.0.0.0` on the host
- Never use `-p 2019:2019` in production (exposes to all interfaces)
- Never expose admin API to the public internet

### Grafana and Prometheus Security

Both Grafana and Prometheus UIs are bound to `127.0.0.1` on the host in production:

```yaml
ports:
  - "127.0.0.1:9090:9090"  # Prometheus - localhost only
  - "127.0.0.1:3000:3000"  # Grafana - localhost only
```

Access remotely via SSH tunneling:
```bash
ssh -L 3000:localhost:3000 -L 9090:localhost:9090 user@production-server
```

## Local Testing

### Prerequisites

1. Docker and Docker Compose installed
2. `.env` file configured with database credentials
3. Database secrets in `database/infra/secrets/`

### Starting the Monitoring Stack Locally

```bash
# Start the full local stack with monitoring
docker compose --profile local up -d

# Or if you want to see logs
docker compose --profile local up
```

This will start:
- API service (port 8080)
- Caddy proxy (ports 8080, 8443, 2019)
- PostgreSQL database
- Prometheus (port 9090)
- Grafana (port 3000)
- PostgreSQL exporter

### Accessing Services Locally

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / (set via GRAFANA_ADMIN_PASSWORD) |
| Prometheus | http://localhost:9090 | None |
| Caddy Admin | http://localhost:2019 | None |
| API | http://localhost:8080 | (your API auth) |

### Verifying the Setup

#### 1. Check that all services are running

```bash
docker ps
```

You should see containers for:
- `oullin_grafana_local`
- `oullin_prometheus_local`
- `oullin_postgres_exporter_local`
- `oullin_local_proxy`
- `oullin_db`
- API container

#### 2. Verify Prometheus is scraping targets

Open http://localhost:9090/targets

All targets should show as "UP":
- caddy (http://caddy_local:2019/metrics)
- postgresql (http://postgres_exporter_local:9187/metrics)
- api (http://api:8080/metrics)
- prometheus (http://localhost:9090/metrics)

#### 3. Test Caddy metrics endpoint

```bash
curl http://localhost:2019/metrics
```

You should see metrics like:
```
caddy_http_requests_total
caddy_http_request_duration_seconds
caddy_http_connections_open
```

#### 4. Test API metrics endpoint

```bash
# From host machine (if API is exposed)
curl http://localhost:8080/metrics

# Or from within a container
docker exec -it oullin_prometheus_local curl http://api:8080/metrics
```

You should see Go runtime metrics like:
```
go_memstats_alloc_bytes
go_goroutines
promhttp_metric_handler_requests_total
```

#### 5. Test PostgreSQL exporter

```bash
docker exec -it oullin_prometheus_local curl http://postgres_exporter_local:9187/metrics
```

You should see database metrics like:
```
pg_stat_database_numbackends
pg_database_size_bytes
pg_stat_database_xact_commit
```

#### 6. Access Grafana Dashboards

1. Open http://localhost:3000
2. Login with `admin` / (your password)
3. Navigate to "Dashboards"
4. You should see three dashboards:
   - **Oullin - Overview**: High-level metrics
   - **PostgreSQL - Database Metrics**: Database performance
   - **Caddy - Proxy Metrics**: Proxy performance

#### 7. Generate Some Traffic

To see metrics populate, generate some API traffic:

```bash
# Make some requests to your API
for i in {1..100}; do
  curl http://localhost:8080/ping
  sleep 0.1
done
```

Then check the dashboards - you should see:
- Request rate increasing in Caddy dashboard
- API memory/goroutines in Overview dashboard
- Database connections in PostgreSQL dashboard

### Common Local Testing Issues

**Problem**: Targets show as "DOWN" in Prometheus

**Solution**:
```bash
# Check container networking
docker network ls
docker network inspect caddy_net

# Restart services
docker compose --profile local restart
```

**Problem**: No metrics appearing in Grafana

**Solution**:
1. Verify Prometheus data source: Grafana → Settings → Data Sources → Prometheus → "Save & Test"
2. Check Prometheus has data: http://localhost:9090/graph
3. Wait 1-2 minutes for initial scraping

**Problem**: Cannot access Grafana

**Solution**:
```bash
# Check Grafana logs
docker logs oullin_grafana_local

# Restart Grafana
docker compose --profile local restart grafana_local
```

### Stopping the Local Stack

```bash
# Stop all services
docker compose --profile local down

# Stop and remove volumes (clean slate)
docker compose --profile local down -v
```

## Production Deployment

### Starting the Production Stack

```bash
# On your production server
docker compose --profile prod up -d
```

### Accessing Services in Production

All services are bound to localhost for security:

| Service | URL (from server) | Access from Local Machine |
|---------|-------------------|---------------------------|
| Grafana | http://localhost:3000 | `ssh -L 3000:localhost:3000 user@server` |
| Prometheus | http://localhost:9090 | `ssh -L 9090:localhost:9090 user@server` |
| Caddy Admin | *(internal network only)* | Not exposed to host for security |

**Note**: The Caddy admin API is only accessible within the Docker network for Prometheus scraping. To access it for debugging, use:
```bash
docker exec -it oullin_proxy_prod curl http://localhost:2019/metrics
```

### Verifying Production Setup

SSH into your server and run:

```bash
# Check Prometheus targets
curl http://localhost:9090/targets

# Check Caddy metrics (from within the container)
docker exec -it oullin_proxy_prod curl http://localhost:2019/metrics

# View Grafana dashboards
# Open SSH tunnel, then access http://localhost:3000 from your browser
```

### Production Monitoring Checklist

- [ ] All Prometheus targets are UP
- [ ] Grafana dashboards are accessible
- [ ] Metrics are being collected (check time series graphs)
- [ ] Alerts are configured (if any)
- [ ] Retention period is appropriate (30 days for prod, 7 days for local)
- [ ] Volumes are backed up regularly

## Useful Prometheus Queries

### API Performance
```promql
# Request rate
rate(promhttp_metric_handler_requests_total[5m])

# Memory usage
go_memstats_alloc_bytes{job="api"}

# Goroutines (check for leaks)
go_goroutines{job="api"}

# GC duration
rate(go_gc_duration_seconds_sum[5m])
```

### Database Performance
```promql
# Active connections
pg_stat_database_numbackends

# Database size growth
delta(pg_database_size_bytes[1h])

# Transaction rate
rate(pg_stat_database_xact_commit[5m])

# Cache hit ratio (should be >90%)
rate(pg_stat_database_blks_hit[5m]) /
(rate(pg_stat_database_blks_hit[5m]) + rate(pg_stat_database_blks_read[5m]))

# Slow queries indicator
rate(pg_stat_database_xact_rollback[5m])
```

### Caddy Performance
```promql
# Request rate by status
sum by(code) (rate(caddy_http_requests_total[5m]))

# 95th percentile response time
histogram_quantile(0.95, rate(caddy_http_request_duration_seconds_bucket[5m]))

# Error rate (5xx responses)
sum(rate(caddy_http_requests_total{code=~"5.."}[5m]))

# Active connections
caddy_http_connections_open
```

## Troubleshooting

### Prometheus Not Scraping

1. Check network connectivity:
   ```bash
   docker exec -it oullin_prometheus_local ping caddy_local
   ```

2. Verify service is exposing metrics:
   ```bash
   docker exec -it oullin_prometheus_local curl http://caddy_local:2019/metrics
   ```

3. Check Prometheus config:
   ```bash
   docker exec -it oullin_prometheus_local cat /etc/prometheus/prometheus.yml
   ```

### High Memory Usage

Monitor memory with:
```bash
docker stats
```

If Prometheus is using too much memory:
- Reduce retention time
- Decrease scrape frequency
- Add more specific metric filters

### Data Not Persisting

Ensure volumes are properly configured:
```bash
docker volume ls
docker volume inspect prometheus_data
docker volume inspect grafana_data
```

## Maintenance

### Backing Up Data

```bash
# Backup Prometheus data
docker run --rm -v prometheus_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/prometheus-backup-$(date +%Y%m%d).tar.gz /data

# Backup Grafana data
docker run --rm -v grafana_data:/data -v $(pwd):/backup alpine \
  tar czf /backup/grafana-backup-$(date +%Y%m%d).tar.gz /data
```

### Updating the Stack

```bash
# Pull latest images
docker compose pull

# Restart with new images
docker compose --profile prod up -d
```

### Cleaning Up Old Data

Prometheus automatically handles retention based on `--storage.tsdb.retention.time` flag.

To manually clean up:
```bash
# Stop Prometheus
docker compose stop prometheus_local

# Clean data
docker run --rm -v prometheus_data:/data alpine rm -rf /data/*

# Restart
docker compose --profile local up -d prometheus_local
```

## Next Steps

1. **Set up Alerting**: Configure Prometheus Alertmanager for critical metrics
2. **Add Custom Metrics**: Instrument your API with custom business metrics
3. **Create Custom Dashboards**: Build dashboards specific to your use case
4. **Export Dashboards**: Share dashboard JSON files with your team

## Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Caddy Metrics](https://caddyserver.com/docs/metrics)
- [PostgreSQL Exporter](https://github.com/prometheus-community/postgres_exporter)
