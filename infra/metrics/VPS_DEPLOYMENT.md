# VPS Deployment Guide

Complete guide for deploying the Oullin monitoring stack on an Ubuntu VPS (Hostinger or similar).

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Server Setup](#initial-server-setup)
3. [Install Docker and Docker Compose](#install-docker-and-docker-compose)
4. [Install Make](#install-make)
5. [Clone Your Repository](#clone-your-repository)
6. [Configure Environment Variables](#configure-environment-variables)
7. [Set Up Docker Secrets](#set-up-docker-secrets)
8. [Configure Firewall](#configure-firewall)
9. [Deploy the Monitoring Stack](#deploy-the-monitoring-stack)
10. [Verify Monitoring Stack](#verify-monitoring-stack)
11. [Access Grafana Remotely](#access-grafana-remotely)
12. [Production Considerations](#production-considerations)
13. [Generate Test Traffic](#generate-test-traffic)
14. [VPS Troubleshooting](#vps-troubleshooting)
15. [Updating the Stack](#updating-the-stack)
16. [Installing Fail2ban](#installing-fail2ban)

---

## Prerequisites

- Hostinger VPS with Ubuntu 20.04 or 22.04 (or similar VPS provider)
- SSH access to your VPS
- Domain name (optional, but recommended for SSL)
- At least 2GB RAM and 20GB storage

---

## Initial Server Setup

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

---

## Install Docker and Docker Compose

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

---

## Install Make

```bash
sudo apt install -y make
```

---

## Clone Your Repository

```bash
cd ~
git clone https://github.com/yourusername/your-repo.git
cd your-repo
```

---

## Configure Environment Variables

Create your `.env` file with production settings:

```bash
cat > .env << 'EOF'
# Database Configuration
POSTGRES_USER=your_db_user
POSTGRES_PASSWORD=your_strong_db_password
POSTGRES_DB=your_database_name

# Grafana Configuration (optional - defaults to "admin")
# Strongly recommended to set a secure password for production
GRAFANA_ADMIN_PASSWORD=your_very_strong_grafana_password

# Production Domain (optional, for SSL)
DOMAIN=your-domain.com

# Environment
ENVIRONMENT=production
EOF
```

**Security Notes:**
- Use strong, unique passwords
- If `GRAFANA_ADMIN_PASSWORD` is not set, it defaults to "admin" - strongly recommended to change for production
- Never commit `.env` to version control
- Consider using a password manager

---

## Set Up Docker Secrets

Avoid piping credentials through `echo` because the literal values end up in your shell history. Use one of the safer patterns below.

### Option 1: Read secrets from secure input

```bash
# Prompt won't echo characters and won't touch shell history
read -s -p "Enter database password: " DB_PASSWORD && echo

echo "$DB_PASSWORD" | docker secret create pg_password - 2>/dev/null || \
  printf "%s" "$DB_PASSWORD" > secrets/pg_password

unset DB_PASSWORD
```

Repeat the same pattern for usernames or other sensitive values you do not want stored on disk.

### Option 2: Write files directly

```bash
mkdir -p secrets
printf "your_db_user" > secrets/pg_username
printf "your_strong_db_password" > secrets/pg_password
printf "your_database_name" > secrets/pg_dbname
chmod 600 secrets/*
```

Store these files somewhere secure (e.g., `pass`, `1Password CLI`, `sops`) and only copy them onto the server when needed.

---

## Configure Firewall

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

---

## Deploy the Monitoring Stack

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

---

## Verify Monitoring Stack

Check Prometheus targets:

```bash
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'
```

All should show `"health": "up"`.

---

## Access Grafana Remotely

From your local machine:

```bash
ssh -L 3000:localhost:3000 deployer@your-vps-ip
```

Then open `http://localhost:3000` in your browser.

**Login:**
- Username: `admin`
- Password: Value from `GRAFANA_ADMIN_PASSWORD`

---

## Production Considerations

### Enable Automatic Backups

Schedule daily backups:

```bash
crontab -e
```

Add:

# NOTE: Update /home/deployer/your-repo to your actual repository path
```cron
# Run daily at 2 AM
0 2 * * * cd /home/deployer/your-repo && make monitor-backup-prod >> /var/log/prometheus-backup.log 2>&1
```

### Monitor Disk Space

```bash
# Check disk usage
df -h

# Check Prometheus data size
docker exec oullin_prometheus du -sh /prometheus
```

### Configure Log Rotation

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

### Enable SSL/TLS (Optional)

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
127.0.0.1:2019 {
    admin {
        metrics
    }
}
```

Caddy will automatically obtain Let's Encrypt certificates.

---

## Generate Test Traffic

```bash
make monitor-traffic-prod
```

Wait a few minutes for data to appear in Grafana.

---

## VPS Troubleshooting

### Services won't start

```bash
# View logs from monitoring services
make monitor-logs         # Local: all services
make monitor-logs-prod    # Production: all services

# Or view individual container logs
docker logs oullin_grafana
docker logs oullin_prometheus

# Check Docker daemon
sudo systemctl status docker
```

### Can't connect via SSH tunnel

```bash
# Verify Grafana is listening
docker exec oullin_grafana netstat -tlnp | grep 3000

# Check if port is already in use locally
lsof -i :3000
```

### Prometheus targets are down

```bash
# Check DNS resolution
docker exec oullin_prometheus nslookup oullin_proxy_prod
docker exec oullin_prometheus nslookup oullin_postgres_exporter

# Verify network
docker network inspect caddy_net oullin_net
```

### Out of disk space

```bash
# Clean up Docker
docker system prune -a --volumes

# Rotate backups (keeps last 5)
make monitor-backup

# Clear old Prometheus data
docker exec oullin_prometheus rm -rf /prometheus/wal/*
```

---

## Updating the Stack

```bash
cd ~/your-repo
git pull origin main

make monitor-down-prod
make monitor-up-prod
```

---

## Installing Fail2ban

```bash
sudo apt install -y fail2ban
sudo systemctl start fail2ban
sudo systemctl enable fail2ban
sudo fail2ban-client status sshd
```

---

## Production Checklist

- ✅ `GRAFANA_ADMIN_PASSWORD` set in `.env` (recommended for production)
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

## Additional Resources

For monitoring-specific documentation, see [README.md](./README.md).
