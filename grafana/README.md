# Grafana Monitoring Dashboard

This directory contains the Grafana configuration for monitoring the Oullin application stack.

## Table of Contents
1. [Access](#access)
2. [Deploying on Ubuntu VPS (Hostinger)](#deploying-on-ubuntu-vps-hostinger)
3. [Pre-configured Dashboards](#pre-configured-dashboards)
4. [Data Source](#data-source)
5. [Creating Custom Dashboards](#creating-custom-dashboards)
6. [Dashboard Best Practices](#dashboard-best-practices)
7. [Directory Structure](#directory-structure)
8. [Example Queries by Service](#example-queries-by-service)
9. [Troubleshooting](#troubleshooting)
10. [Resources](#resources)
11. [Quick Reference](#quick-reference)

---

## Access

Grafana is accessible at `http://localhost:3000` (from the server)

**Default Credentials:**
- Username: `admin`
- Password: Set via `GRAFANA_ADMIN_PASSWORD` environment variable (required in `.env` file)

**Security Note:** The `GRAFANA_ADMIN_PASSWORD` environment variable is required and must be set in your `.env` file. Do not use default passwords.

### Remote Access

To access Grafana from your local machine:

```bash
ssh -L 3000:localhost:3000 user@your-server.com
```

Then open `http://localhost:3000` in your browser.

---

## Deploying on Ubuntu VPS (Hostinger)

This guide walks you through deploying the full monitoring stack (Prometheus, Grafana, postgres_exporter, Caddy) on an Ubuntu VPS from Hostinger.

### Prerequisites

- Hostinger VPS with Ubuntu 20.04 or 22.04
- SSH access to your VPS
- Domain name (optional, but recommended for SSL)
- At least 2GB RAM and 20GB storage

### Step 1: Initial Server Setup

Connect to your VPS via SSH:

```bash
ssh root@your-vps-ip
```

Update the system:

```bash
apt update && apt upgrade -y
```

Create a non-root user (recommended for security):

```bash
# Create user
adduser deployer

# Add to sudo group
usermod -aG sudo deployer

# Switch to new user
su - deployer
```

### Step 2: Install Docker and Docker Compose

Install required packages:

```bash
sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
```

Add Docker's official GPG key:

```bash
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
```

Add Docker repository:

```bash
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

Install Docker:

```bash
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

Add your user to the docker group:

```bash
sudo usermod -aG docker ${USER}
```

Log out and back in for group changes to take effect, then verify:

```bash
docker --version
docker compose version
```

### Step 3: Install Make (if not present)

```bash
sudo apt install -y make
```

### Step 4: Clone Your Repository

```bash
cd ~
git clone https://github.com/yourusername/your-repo.git
cd your-repo
```

### Step 5: Configure Environment Variables

Create your `.env` file with production settings:

```bash
cat > .env << 'EOF'
# Database Configuration
POSTGRES_USER=your_db_user
POSTGRES_PASSWORD=your_strong_db_password
POSTGRES_DB=your_database_name

# Grafana Configuration (REQUIRED - no default)
GRAFANA_ADMIN_PASSWORD=your_very_strong_grafana_password

# Production Domain (optional, for SSL)
DOMAIN=your-domain.com

# Environment
ENVIRONMENT=production
EOF
```

**Important Security Notes:**
- Use strong, unique passwords for all credentials
- Never commit `.env` to version control (already in `.gitignore`)
- Consider using a password manager to generate strong passwords

### Step 6: Set Up Docker Secrets

Create Docker secrets for sensitive data:

```bash
# Create secrets directory (if using file-based secrets for local testing)
mkdir -p secrets

# PostgreSQL credentials
echo "your_db_user" | docker secret create pg_username - 2>/dev/null || \
  echo "your_db_user" > secrets/pg_username

echo "your_strong_db_password" | docker secret create pg_password - 2>/dev/null || \
  echo "your_strong_db_password" > secrets/pg_password

echo "your_database_name" | docker secret create pg_dbname - 2>/dev/null || \
  echo "your_database_name" > secrets/pg_dbname
```

**Note:** Docker secrets work differently in Swarm mode vs Compose mode. The above creates file-based secrets for Compose.

### Step 7: Configure Firewall

Set up UFW firewall to secure your VPS:

```bash
# Enable UFW
sudo ufw --force enable

# Allow SSH (IMPORTANT: Do this first!)
sudo ufw allow 22/tcp

# Allow HTTP and HTTPS (for Caddy)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Verify rules
sudo ufw status
```

**Do NOT expose Prometheus (9090), Grafana (3000), or postgres_exporter (9187) ports directly.** Access these services via SSH tunnel for security.

### Step 8: Deploy the Monitoring Stack

Deploy using the production profile:

```bash
# Start the monitoring stack with production profile
make monitor-up-prod

# Or manually:
docker compose --profile prod up -d
```

Verify all services are running:

```bash
docker compose ps
```

You should see:
- `oullin_prometheus` - Running
- `oullin_grafana` - Running
- `oullin_postgres_exporter` - Running
- `oullin_proxy_prod` (Caddy) - Running
- `oullin_db` (PostgreSQL) - Running

### Step 9: Verify Monitoring Stack

Check that Prometheus is scraping targets:

```bash
# From your VPS
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

All targets should show `"health": "up"`.

### Step 10: Access Grafana Remotely

Create an SSH tunnel from your local machine to access Grafana securely:

```bash
# From your LOCAL machine (not the VPS)
ssh -L 3000:localhost:3000 deployer@your-vps-ip
```

Then open `http://localhost:3000` in your browser.

**Login:**
- Username: `admin`
- Password: The value you set in `GRAFANA_ADMIN_PASSWORD`

### Step 11: Production Considerations

#### Enable Automatic Restarts

Ensure containers restart automatically:

```bash
# Check restart policies
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.RestartPolicy}}"
```

The `docker-compose.yml` should have `restart: unless-stopped` for all services.

#### Set Up Backups

Schedule regular Prometheus data backups:

```bash
# Create a cron job for daily backups
crontab -e
```

Add this line to backup daily at 2 AM:

```cron
0 2 * * * cd /home/deployer/your-repo && make monitor-backup >> /var/log/prometheus-backup.log 2>&1
```

#### Monitor Disk Space

Prometheus data can grow over time. Monitor disk usage:

```bash
# Check disk space
df -h

# Check Prometheus data size
docker exec oullin_prometheus du -sh /prometheus
```

Consider setting up retention policies in `prometheus/prometheus.yml`:

```yaml
global:
  # Keep data for 30 days
  storage.tsdb.retention.time: 30d
```

#### Configure Log Rotation

Set up log rotation for Docker containers:

```bash
sudo tee /etc/docker/daemon.json > /dev/null << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

# Restart Docker
sudo systemctl restart docker

# Restart containers
make monitor-restart-prod
```

#### Enable SSL/TLS (Optional)

If you have a domain, configure Caddy for automatic HTTPS:

Edit `caddy/Caddyfile.prod`:

```caddyfile
your-domain.com {
    reverse_proxy api:8080

    log {
        output file /var/log/caddy/access.log
    }
}

# Admin API (internal only)
:2019 {
    admin {
        metrics
    }
}
```

Caddy will automatically obtain Let's Encrypt certificates.

### Step 12: Generate Test Traffic

Generate some traffic to populate the dashboards:

```bash
# From the VPS
make monitor-traffic-prod
```

Wait a few minutes for data to appear in Grafana.

### Troubleshooting VPS Deployment

#### Services won't start

```bash
# Check logs
make monitor-logs-grafana
make monitor-logs-prometheus

# Check Docker daemon
sudo systemctl status docker
```

#### Can't connect via SSH tunnel

```bash
# Verify Grafana is listening
docker exec oullin_grafana netstat -tlnp | grep 3000

# Check if port is already in use locally
lsof -i :3000
```

#### Prometheus targets are down

```bash
# Check container DNS resolution
docker exec oullin_prometheus nslookup oullin_proxy_prod
docker exec oullin_prometheus nslookup oullin_postgres_exporter

# Verify containers are on the same network
docker network inspect your-repo_default
```

#### Out of disk space

```bash
# Clean up Docker resources
docker system prune -a --volumes

# Rotate old backups
make monitor-backup  # This automatically keeps only last 5 backups

# Clear old Prometheus data (if retention is too long)
docker exec oullin_prometheus rm -rf /prometheus/wal/*
```

### Updating the Monitoring Stack

To update your monitoring stack:

```bash
# Pull latest changes
cd ~/your-repo
git pull origin main

# Rebuild and restart
make monitor-down-prod
make monitor-up-prod

# Or with Docker Compose directly
docker compose --profile prod down
docker compose --profile prod up -d --build
```

### Monitoring Resource Usage

Keep an eye on VPS resource usage:

```bash
# CPU and Memory usage
docker stats

# Disk usage by container
docker system df -v

# Container logs size
sudo du -sh /var/lib/docker/containers/*/*-json.log
```

### Security Checklist

- ✅ Firewall configured (UFW)
- ✅ Only necessary ports exposed (22, 80, 443)
- ✅ Monitoring services NOT exposed to internet
- ✅ Strong passwords for all services
- ✅ Docker secrets for sensitive data
- ✅ Regular backups scheduled
- ✅ Log rotation configured
- ✅ SSH key-based authentication (recommended)
- ✅ Fail2ban installed (optional but recommended)

### Installing Fail2ban (Recommended)

Protect against brute-force SSH attacks:

```bash
sudo apt install -y fail2ban

# Start and enable
sudo systemctl start fail2ban
sudo systemctl enable fail2ban

# Check status
sudo fail2ban-client status sshd
```

---

## Pre-configured Dashboards

Three dashboards are automatically provisioned:

### 1. Oullin - Overview
**File:** `oullin-overview-oullin-overview.json`

High-level view of all services:
- Caddy request rate
- PostgreSQL active connections
- HTTP requests by status code
- API memory usage and goroutines

### 2. PostgreSQL - Database Metrics
**File:** `oullin-postgresql-postgresql-database-metrics.json`

Detailed database monitoring:
- Active connections
- Database size
- Transaction rates
- Database operations (inserts, updates, deletes)
- Cache hit ratio
- Lock statistics

### 3. Caddy - Proxy Metrics
**File:** `oullin-caddy-caddy-proxy-metrics.json`

Reverse proxy performance:
- Total request rate
- Response time percentiles (p50, p95, p99)
- Requests by status code
- Traffic rate (request/response sizes)
- Request errors

---

## Data Source

Grafana is pre-configured with Prometheus as the default data source, automatically connecting to the Prometheus service at `http://prometheus:9090`.

---

## Creating Custom Dashboards

### Method 1: Create in UI and Export (Recommended)

This is the easiest approach for creating custom dashboards.

#### Step 1: Start Grafana

```bash
make monitor-up
make monitor-grafana  # Opens http://localhost:3000
```

Login: `admin` / (your GRAFANA_ADMIN_PASSWORD)

#### Step 2: Create a New Dashboard

1. Click **"+"** → **"Dashboard"** → **"Add visualization"**
2. Select **"Prometheus"** as the data source
3. Write your PromQL query:
   ```promql
   # Example queries
   rate(caddy_http_request_count_total[5m])
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

#### Step 3: Export Dashboard (Manual)

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

#### Step 4: Export Dashboard (Automated)

Use the export script:

```bash
make monitor-export-dashboards
```

This will:
1. List all dashboards in Grafana
2. Let you select which to export
3. Save to `grafana/dashboards/`
4. Format properly for provisioning

#### Step 5: Reload Grafana

```bash
make monitor-restart
```

Your dashboard will now auto-load on startup!

---

### Method 2: Use Community Dashboards

Grafana has thousands of pre-built dashboards at https://grafana.com/grafana/dashboards/

#### Popular Dashboards for Our Stack:

**PostgreSQL:**
- [9628](https://grafana.com/grafana/dashboards/9628) - PostgreSQL Database
- [455](https://grafana.com/grafana/dashboards/455) - PostgreSQL Stats

**Go Applications:**
- [10826](https://grafana.com/grafana/dashboards/10826) - Go Metrics
- [6671](https://grafana.com/grafana/dashboards/6671) - Go Processes

**Caddy:**
- Community dashboards for reverse proxies work well

#### How to Import:

**Via Grafana UI:**
1. Click **"+"** → **"Import"**
2. Enter dashboard ID (e.g., `9628`)
3. Click **"Load"**
4. Select **"Prometheus"** as data source
5. Click **"Import"**

**Via Dashboard JSON:**
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

### Method 3: Generate with Grafonnet (Advanced)

Grafonnet is a Jsonnet library for generating Grafana dashboards programmatically.

#### Why Use Grafonnet?
- Generate multiple similar dashboards
- Version control dashboard logic, not JSON
- Template dashboards with variables
- Consistent styling across all dashboards

#### Example Grafonnet Dashboard:

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

#### Generate JSON:

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

### Method 4: Edit Existing JSON

You can directly edit dashboard JSON files, but this requires understanding the schema.

#### Dashboard JSON Structure:

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

#### Key Properties:

- **id**: Must be `null` for provisioned dashboards
- **uid**: Unique identifier (optional for provisioned)
- **panels**: Array of visualization panels
- **gridPos**: Position and size (x, y, w, h) in grid units
- **targets**: Prometheus queries
- **overwrite**: `true` to replace existing dashboard

#### Tips for Editing:

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

## Directory Structure

```text
grafana/
├── README.md
├── dashboards/               # Dashboard JSON files
│   ├── oullin-overview-oullin-overview.json
│   ├── oullin-postgresql-postgresql-database-metrics.json
│   └── oullin-caddy-caddy-proxy-metrics.json
├── scripts/
│   └── export-dashboards.sh # Dashboard export script
└── provisioning/
    ├── datasources/          # Data source configuration
    │   └── prometheus.yml
    └── dashboards/           # Dashboard provisioning config
        └── default.yml
```

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
sum by(code) (rate(caddy_http_request_count_total[5m]))

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

1. Check JSON syntax: `jq . < grafana/dashboards/my-dashboard.json`
2. Ensure `"id": null` in dashboard definition
3. Check Grafana logs: `docker logs oullin_grafana` or `make monitor-logs-grafana`
4. Verify file is in correct directory
5. Verify Prometheus connection: Settings → Data Sources → Prometheus → "Save & Test"
6. Ensure Prometheus is running: `docker ps | grep prometheus`

### No Data in Panels

1. Verify Prometheus is scraping targets: `http://localhost:9090/targets` or `make monitor-targets`
2. Test query in Prometheus: `http://localhost:9090`
3. Verify data source in panel settings
4. Check time range isn't too far in past
5. Check that services are exposing metrics
6. Wait a few minutes for initial data collection

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
# Start monitoring stack
make monitor-up

# Open Grafana in browser
make monitor-grafana

# Export existing dashboards
make monitor-export-dashboards

# View current dashboard files
ls -la grafana/dashboards/

# Test a PromQL query
curl 'http://localhost:9090/api/v1/query?query=up'

# Restart to load new dashboards
make monitor-restart

# View Grafana logs
make monitor-logs-grafana

# Check Prometheus targets
make monitor-targets
```
