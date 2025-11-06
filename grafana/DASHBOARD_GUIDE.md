# Grafana Dashboard Creation Guide

This guide explains how to create, export, and manage Grafana dashboards for the Oullin monitoring stack.

## Table of Contents
1. [Current Dashboards](#current-dashboards)
2. [Method 1: Create in UI and Export (Recommended)](#method-1-create-in-ui-and-export-recommended)
3. [Method 2: Use Community Dashboards](#method-2-use-community-dashboards)
4. [Method 3: Generate with Grafonnet (Advanced)](#method-3-generate-with-grafonnet-advanced)
5. [Method 4: Edit Existing JSON](#method-4-edit-existing-json)
6. [Dashboard Best Practices](#dashboard-best-practices)

---

## Current Dashboards

The project includes three pre-configured dashboards:

1. **overview.json** - High-level metrics from all services
2. **postgresql.json** - Detailed database monitoring
3. **caddy.json** - Reverse proxy performance

These were manually created to provide a starting point.

---

## Method 1: Create in UI and Export (Recommended)

This is the easiest approach for creating custom dashboards.

### Step 1: Start Grafana

```bash
make monitor-up
make monitor-grafana  # Opens http://localhost:3000
```

Login: `admin` / (your GRAFANA_ADMIN_PASSWORD)

### Step 2: Create a New Dashboard

1. Click **"+"** → **"Dashboard"** → **"Add visualization"**
2. Select **"Prometheus"** as the data source
3. Write your PromQL query:
   ```promql
   # Example queries
   rate(caddy_http_requests_total[5m])
   go_memstats_alloc_bytes{job="api"}
   pg_stat_database_numbackends
   ```
4. Choose visualization type:
   - **Time series** - For trends over time
   - **Stat** - For single current values
   - **Gauge** - For percentage/threshold values
   - **Table** - For tabular data

5. Configure panel:
   - **Panel title**: Descriptive name
   - **Description**: What the panel shows
   - **Unit**: bytes, requests/sec, percent, etc.
   - **Thresholds**: Warning/critical levels
   - **Legend**: Show/hide, placement

6. Add more panels by clicking **"Add"** → **"Visualization"**

7. Arrange panels by dragging them

8. Save dashboard: Click **"Save dashboard"** icon (top right)

### Step 3: Export Dashboard (Manual)

1. Open your dashboard
2. Click the **"Share"** icon (top right)
3. Go to **"Export"** tab
4. **Option A**: Click **"Save to file"** - downloads JSON
5. **Option B**: Click **"View JSON"** - copy the JSON

6. Save to project:
   ```bash
   # Replace MY-DASHBOARD with your filename
   cat > ./grafana/dashboards/my-custom-dashboard.json << 'EOF'
   {
     paste your JSON here
   }
   EOF
   ```

### Step 4: Export Dashboard (Automated)

Use the export script:

```bash
make monitor-export-dashboards
```

This will:
1. List all dashboards in Grafana
2. Let you select which to export
3. Save to `grafana/dashboards/`
4. Format properly for provisioning

### Step 5: Reload Grafana

```bash
make monitor-restart
```

Your dashboard will now auto-load on startup!

---

## Method 2: Use Community Dashboards

Grafana has thousands of pre-built dashboards at https://grafana.com/grafana/dashboards/

### Popular Dashboards for Our Stack:

**PostgreSQL:**
- [9628](https://grafana.com/grafana/dashboards/9628) - PostgreSQL Database
- [455](https://grafana.com/grafana/dashboards/455) - PostgreSQL Stats

**Go Applications:**
- [10826](https://grafana.com/grafana/dashboards/10826) - Go Metrics
- [6671](https://grafana.com/grafana/dashboards/6671) - Go Processes

**Caddy:**
- Community dashboards for reverse proxies work well

### How to Import:

#### Via Grafana UI:
1. Click **"+"** → **"Import"**
2. Enter dashboard ID (e.g., `9628`)
3. Click **"Load"**
4. Select **"Prometheus"** as data source
5. Click **"Import"**

#### Via Dashboard JSON:
1. Visit dashboard page (e.g., https://grafana.com/grafana/dashboards/9628)
2. Click **"Download JSON"**
3. Save to `grafana/dashboards/postgres-community.json`
4. Edit the file and add these properties:
   ```json
   {
     "dashboard": { ... existing content ... },
     "overwrite": true,
     "inputs": [
       {
         "name": "DS_PROMETHEUS",
         "type": "datasource",
         "pluginId": "prometheus",
         "value": "Prometheus"
       }
     ]
   }
   ```
5. Restart Grafana: `make monitor-restart`

---

## Method 3: Generate with Grafonnet (Advanced)

Grafonnet is a Jsonnet library for generating Grafana dashboards programmatically.

### Why Use Grafonnet?
- Generate multiple similar dashboards
- Version control dashboard logic, not JSON
- Template dashboards with variables
- Consistent styling across all dashboards

### Example Grafonnet Dashboard:

Create `grafana/grafonnet/api-metrics.jsonnet`:

```jsonnet
local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local prometheus = grafana.prometheus;
local graphPanel = grafana.graphPanel;

dashboard.new(
  'API Metrics',
  schemaVersion=16,
  tags=['oullin', 'api'],
  time_from='now-6h',
)
.addPanel(
  graphPanel.new(
    'Request Rate',
    datasource='Prometheus',
    span=6,
  )
  .addTarget(
    prometheus.target(
      'rate(promhttp_metric_handler_requests_total[5m])',
      legendFormat='{{code}}',
    )
  ),
  gridPos={x: 0, y: 0, w: 12, h: 8}
)
.addPanel(
  graphPanel.new(
    'Memory Usage',
    datasource='Prometheus',
    span=6,
  )
  .addTarget(
    prometheus.target(
      'go_memstats_alloc_bytes',
      legendFormat='Allocated',
    )
  ),
  gridPos={x: 12, y: 0, w: 12, h: 8}
)
```

### Generate JSON:

```bash
# Install jsonnet
go install github.com/google/go-jsonnet/cmd/jsonnet@latest

# Install grafonnet
git clone https://github.com/grafana/grafonnet-lib.git grafana/grafonnet-lib

# Generate dashboard
jsonnet -J grafana/grafonnet-lib grafana/grafonnet/api-metrics.jsonnet \
  > grafana/dashboards/api-metrics-generated.json
```

---

## Method 4: Edit Existing JSON

You can directly edit dashboard JSON files, but this requires understanding the schema.

### Dashboard JSON Structure:

```json
{
  "dashboard": {
    "title": "My Dashboard",
    "tags": ["oullin", "monitoring"],
    "timezone": "browser",
    "schemaVersion": 39,
    "panels": [
      {
        "id": 1,
        "type": "timeseries",
        "title": "Panel Title",
        "gridPos": {"x": 0, "y": 0, "w": 12, "h": 8},
        "datasource": {"type": "prometheus", "uid": "prometheus"},
        "targets": [
          {
            "expr": "rate(metric_name[5m])",
            "legendFormat": "{{label}}",
            "refId": "A"
          }
        ]
      }
    ]
  },
  "overwrite": true
}
```

### Key Properties:

- **id**: Must be `null` for provisioned dashboards
- **uid**: Unique identifier (optional for provisioned)
- **panels**: Array of visualization panels
- **gridPos**: Position and size (x, y, w, h) in grid units
- **targets**: Prometheus queries
- **overwrite**: `true` to replace existing dashboard

### Tips for Editing:

1. **Copy an existing dashboard** as a template
2. **Use a JSON formatter** for readability
3. **Validate JSON** before saving
4. **Test in Grafana UI** before committing

---

## Dashboard Best Practices

### 1. Organization
- **One dashboard per service** (API, Database, Proxy)
- **Overview dashboard** for high-level metrics
- **Detail dashboards** for deep dives
- Use **tags** for categorization

### 2. Panel Design
- **Clear titles** that explain what's shown
- **Descriptions** for complex metrics
- **Consistent colors** across dashboards
- **Appropriate units** (bytes, %, req/s)
- **Thresholds** for warnings/errors

### 3. Query Performance
- **Avoid high-cardinality labels** in queries
- **Use recording rules** for expensive queries
- **Limit time range** to what's needed
- **Use `rate()`** instead of raw counters

### 4. Layout
- **Most important metrics** at the top
- **Related metrics** grouped together
- **Consistent panel sizes** for clean look
- **Use rows** to organize sections

### 5. Variables (Advanced)
Add template variables for filtering:
- **Environment** (local, staging, production)
- **Service** (api, database, caddy)
- **Time range** picker

Example variable:
```json
"templating": {
  "list": [
    {
      "name": "environment",
      "type": "custom",
      "options": ["local", "production"],
      "current": {"text": "local", "value": "local"}
    }
  ]
}
```

Use in query: `metric_name{environment="$environment"}`

---

## Example Queries by Service

### API Metrics (Go Application)

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

# Active connections
caddy_http_connections_open

# Traffic rate
rate(caddy_http_response_size_bytes_sum[5m])
```

---

## Troubleshooting

### Dashboard Not Loading

1. Check JSON syntax: `jq . < grafana/dashboards/my-dashboard.json`
2. Ensure `"id": null` in dashboard definition
3. Check Grafana logs: `make monitor-logs-grafana`
4. Verify file is in correct directory

### No Data in Panels

1. Check Prometheus is scraping: `make monitor-targets`
2. Test query in Prometheus: http://localhost:9090
3. Verify data source in panel settings
4. Check time range isn't too far in past

### Dashboard Not Auto-Loading

1. Verify provisioning config: `grafana/provisioning/dashboards/default.yml`
2. Check file permissions: `ls -la grafana/dashboards/`
3. Restart Grafana: `make monitor-restart`
4. Check mount in docker-compose: `./grafana/dashboards:/var/lib/grafana/dashboards:ro`

---

## Resources

- [Grafana Dashboard Documentation](https://grafana.com/docs/grafana/latest/dashboards/)
- [Prometheus Query Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Community Dashboards](https://grafana.com/grafana/dashboards/)
- [Grafonnet Library](https://github.com/grafana/grafonnet-lib)

---

## Quick Reference

```bash
# Start monitoring
make monitor-up

# Export existing dashboards
make monitor-export-dashboards

# View current dashboards
ls -la grafana/dashboards/

# Test a PromQL query
curl 'http://localhost:9090/api/v1/query?query=up'

# Restart to load new dashboards
make monitor-restart

# Open Grafana
make monitor-grafana
```
