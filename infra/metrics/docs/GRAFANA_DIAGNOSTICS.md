# Grafana Caddy Metrics Troubleshooting Guide

## Issue
Grafana dashboards showing no data for Caddy metrics in production:
- `[oullin-overview]`: Caddy Request Rate shows no data
- `[caddy-proxy-metrics]`: All Caddy panels show no data

## Diagnostic Steps (Run on Production Server)

### 1. Check if services are running
```bash
docker ps --filter "name=oullin" --format "table {{.Names}}\t{{.Status}}"
```

Expected: `oullin_proxy_prod`, `oullin_prometheus`, `oullin_grafana` should all be running.

### 2. Verify Caddy is exposing metrics
```bash
# Check if Caddy metrics endpoint is accessible from Prometheus container
docker exec oullin_prometheus wget -qO- http://caddy_prod:9180/metrics | head -20
```

Expected: You should see Prometheus-format metrics starting with `caddy_`.

### 3. Check Prometheus targets
```bash
# Check if Prometheus can scrape Caddy
docker exec oullin_prometheus wget -qO- http://localhost:9090/api/v1/targets | grep caddy -A 10
```

Expected: `"health":"up"` for the caddy target.

Alternative (from host):
```bash
curl -s http://localhost:9090/api/v1/targets | grep caddy -A 10
```

### 4. Verify actual Caddy metric names
```bash
# See what metrics Caddy is actually exposing
docker exec oullin_prometheus wget -qO- http://caddy_prod:9180/metrics | grep "^caddy_http" | cut -d'{' -f1 | sort -u
```

This shows the actual metric names available.

### 5. Test if Prometheus has scraped any Caddy data
```bash
# Query Prometheus for Caddy metrics
curl -s 'http://localhost:9090/api/v1/query?query=up{job="caddy"}'
```

Expected: `"value":[<timestamp>,"1"]` indicating the target is up.

### 6. Check if specific metrics exist
```bash
# Check if the exact metrics the dashboard queries exist
curl -s 'http://localhost:9090/api/v1/query?query=caddy_http_request_duration_seconds_count'
curl -s 'http://localhost:9090/api/v1/query?query=caddy_http_response_size_bytes_sum'
curl -s 'http://localhost:9090/api/v1/query?query=caddy_http_request_errors_total'
```

### 7. Check Prometheus logs
```bash
docker logs oullin_prometheus --tail 100 | grep -i "caddy\|error"
```

## Quick Fix Commands

If services are running but metrics aren't showing:

```bash
# Restart monitoring stack
make monitor-restart-prod

# Or manually:
docker compose --profile prod restart prometheus grafana

# Generate some traffic to populate metrics
make monitor-traffic-prod
```

## Known Issues & Fixes

### Issue: Metric names don't match

If step 4 shows different metric names than what the dashboard expects, you'll need to update the Grafana dashboard queries.

**Common Caddy v2 metrics:**
- `caddy_http_requests_total` - Total requests counter
- `caddy_http_request_duration_seconds_bucket` - Request duration histogram
- `caddy_http_request_duration_seconds_count` - Request count
- `caddy_http_request_duration_seconds_sum` - Total request duration

**Metrics that may NOT exist:**
- `caddy_http_response_size_bytes_sum` - Not in standard Caddy metrics
- `caddy_http_request_errors_total` - Not in standard Caddy metrics

### Issue: Prometheus can't reach Caddy

If Prometheus shows target as DOWN:

1. Verify both containers are on `caddy_net`:
```bash
docker inspect oullin_prometheus | grep -A 10 Networks
docker inspect oullin_proxy_prod | grep -A 10 Networks
```

2. Test network connectivity:
```bash
docker exec oullin_prometheus ping -c 2 caddy_prod
docker exec oullin_prometheus wget -qO- http://caddy_prod:9180/metrics
```

### Issue: No traffic, so no metrics

Caddy metrics only appear after handling requests:

```bash
# Generate test traffic (run from production server)
for i in {1..50}; do curl -s http://localhost/api/health > /dev/null; done
```

## Solution: Update Dashboard Queries

Based on actual metrics available, update the Grafana dashboard JSON files to use correct metric names.

Current dashboard location: `infra/metrics/grafana/dashboards/`

### Example Fix for Traffic Rate Panel

Replace:
```promql
rate(caddy_http_response_size_bytes_sum[5m])
```

With:
```promql
rate(caddy_http_request_duration_seconds_count[5m])
```

### Example Fix for Request Errors Panel

Replace:
```promql
sum(rate(caddy_http_request_errors_total[5m])) or vector(0)
```

With:
```promql
sum(rate(caddy_http_requests_total{code=~"5.."}[5m])) or vector(0)
```

## Expected Behavior

After fixes:
1. Prometheus target `caddy_prod:9180` shows as UP
2. Dashboard panels populate with data after traffic
3. Metrics update every 15s (scrape interval)
