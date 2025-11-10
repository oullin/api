# Monitoring Stack Documentation

Complete guide for deploying, managing, and monitoring the Oullin application stack with Prometheus, Grafana, and related tools.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Security Model](#security-model)
4. [Deploying on Ubuntu VPS (Hostinger)](#deploying-on-ubuntu-vps-hostinger)
5. [Grafana Dashboards](#grafana-dashboards)
6. [Creating Custom Dashboards](#creating-custom-dashboards)
7. [Prometheus Queries](#prometheus-queries)
8. [Troubleshooting](#troubleshooting)
9. [Maintenance & Backup](#maintenance--backup)
10. [Resources](#resources)

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

---

## Quick Start

### Local Development

**Prerequisites:**
- Docker and Docker Compose installed
- `.env` file with `GRAFANA_ADMIN_PASSWORD` set (required - no default)
- Database secrets in `database/infra/secrets/`

**Setup:**

```bash
# 1. Set Grafana admin password in .env file
echo "GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 32)" >> .env

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
   - Production: Only accessible within Docker network
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

## Deploying on Ubuntu VPS (Hostinger)

Complete guide for deploying the monitoring stack on a Hostinger Ubuntu VPS.

### Prerequisites

- Hostinger VPS with Ubuntu 20.04 or 22.04
- SSH access to your VPS
- Domain name (optional, but recommended for SSL)
- At least 2GB RAM and 20GB storage

### Step 1: Initial Server Setup

Connect to your VPS:

```bash
ssh root@your-vps-ip
```

Update the system:

```bash
apt update && apt upgrade -y
```

Create a non-root user:

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

Log out and back in, then verify:

```bash
docker --version
docker compose version
```

### Step 3: Install Make

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

**Security Notes:**
- Use strong, unique passwords
- Never commit `.env` to version control
- Consider using a password manager

### Step 6: Set Up Docker Secrets

Create Docker secrets:

```bash
# Create secrets directory
mkdir -p secrets

# PostgreSQL credentials
echo "your_db_user" | docker secret create pg_username - 2>/dev/null || \
  echo "your_db_user" > secrets/pg_username

echo "your_strong_db_password" | docker secret create pg_password - 2>/dev/null || \
  echo "your_strong_db_password" > secrets/pg_password

echo "your_database_name" | docker secret create pg_dbname - 2>/dev/null || \
  echo "your_database_name" > secrets/pg_dbname
```

### Step 7: Configure Firewall

Set up UFW:

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

**Do NOT expose Prometheus (9090), Grafana (3000), or postgres_exporter (9187) ports!**

### Step 8: Deploy the Monitoring Stack

```bash
# Start with production profile
make monitor-up-prod
# Or: docker compose --profile prod up -d
```

Verify services:

```bash
docker compose ps
```

Expected containers:
- `oullin_prometheus`
- `oullin_grafana`
- `oullin_postgres_exporter`
- `oullin_proxy_prod`
- `oullin_db`

### Step 9: Verify Monitoring Stack

Check Prometheus targets:

```bash
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

All should show `"health": "up"`.

### Step 10: Access Grafana Remotely

From your local machine:

```bash
ssh -L 3000:localhost:3000 deployer@your-vps-ip
```

Then open `http://localhost:3000` in your browser.

**Login:**
- Username: `admin`
- Password: Value from `GRAFANA_ADMIN_PASSWORD`

### Step 11: Production Considerations

#### Enable Automatic Backups

Schedule daily backups:

```bash
crontab -e
```

Add:

```cron
0 2 * * * cd /home/deployer/your-repo && make monitor-backup >> /var/log/prometheus-backup.log 2>&1
```

#### Monitor Disk Space

```bash
# Check disk usage
df -h

# Check Prometheus data size
docker exec oullin_prometheus du -sh /prometheus
```

#### Configure Log Rotation

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

sudo systemctl restart docker
make monitor-restart-prod
```

#### Enable SSL/TLS (Optional)

If you have a domain, configure Caddy for automatic HTTPS.

Edit `infra/caddy/Caddyfile.prod`:

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

```bash
make monitor-traffic-prod
```

Wait a few minutes for data to appear in Grafana.

### VPS Troubleshooting

#### Services won't start

```bash
make monitor-logs-grafana
make monitor-logs-prometheus
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
# Check DNS resolution
docker exec oullin_prometheus nslookup oullin_proxy_prod
docker exec oullin_prometheus nslookup oullin_postgres_exporter

# Verify network
docker network inspect your-repo_default
```

#### Out of disk space

```bash
# Clean up Docker
docker system prune -a --volumes

# Rotate backups (keeps last 5)
make monitor-backup

# Clear old Prometheus data
docker exec oullin_prometheus rm -rf /prometheus/wal/*
```

### Updating the Stack

```bash
cd ~/your-repo
git pull origin main

make monitor-down-prod
make monitor-up-prod
```

### Installing Fail2ban (Recommended)

```bash
sudo apt install -y fail2ban
sudo systemctl start fail2ban
sudo systemctl enable fail2ban
sudo fail2ban-client status sshd
```

---

## Grafana Dashboards

### Accessing Dashboards

**Local:** http://localhost:3000
**Production:** SSH tunnel then http://localhost:3000

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

Grafana has thousands of pre-built dashboards at https://grafana.com/grafana/dashboards/

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

```bash
# Check JSON syntax
jq . < infra/metrics/grafana/dashboards/my-dashboard.json

# Check Grafana logs
docker logs oullin_grafana
make monitor-logs-grafana

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
docker volume inspect prometheus_data
docker volume inspect grafana_data
```

---

## Maintenance & Backup

### Backing Up Data

**Automated backup** (recommended):

```bash
# Runs daily via cron, keeps last 5 backups
make monitor-backup
```

Backups saved to: `storage/monitoring/backups/prometheus-backup-YYYYMMDD-HHMMSS.tar.gz`

**Manual backup:**

```bash
# Backup Prometheus data
docker run --rm -v prometheus_data:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/prometheus-backup-$(date +%Y%m%d-%H%M%S).tar.gz /data

# Backup Grafana data
docker run --rm -v grafana_data:/data -v $(pwd)/backups:/backup alpine \
  tar czf /backup/grafana-backup-$(date +%Y%m%d-%H%M%S).tar.gz /data
```

### Restoring from Backup

```bash
# Stop services
make monitor-down

# Restore Prometheus data
docker run --rm -v prometheus_data:/data -v $(pwd)/backups:/backup alpine \
  sh -c "rm -rf /data/* && tar xzf /backup/prometheus-backup-YYYYMMDD-HHMMSS.tar.gz -C /"

# Restore Grafana data
docker run --rm -v grafana_data:/data -v $(pwd)/backups:/backup alpine \
  sh -c "rm -rf /data/* && tar xzf /backup/grafana-backup-YYYYMMDD-HHMMSS.tar.gz -C /"

# Restart services
make monitor-up
```

### Updating the Stack

```bash
# Pull latest images
docker compose pull

# Restart with new images
make monitor-restart
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
docker run --rm -v prometheus_data:/data alpine rm -rf /data/*

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
make monitor-logs-grafana
make monitor-logs-prometheus

# Maintenance
make monitor-backup          # Backup Prometheus data
make monitor-restart         # Restart services
make monitor-export-dashboards

# Cleanup
make monitor-down            # Stop services
make monitor-clean           # Clean up data
```

### Production Checklist

- ✅ `GRAFANA_ADMIN_PASSWORD` set in `.env`
- ✅ Firewall configured (UFW)
- ✅ Services bound to localhost
- ✅ SSH tunneling configured
- ✅ Backups scheduled (cron)
- ✅ Log rotation configured
- ✅ SSL/TLS enabled (if domain)
- ✅ Fail2ban installed
- ✅ All Prometheus targets UP
- ✅ Dashboards accessible
- ✅ Retention policies set
- ✅ Volumes backed up regularly

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
