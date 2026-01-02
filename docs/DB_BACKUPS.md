# Database Backups
> db-backup.sh

A comprehensive PostgreSQL backup and restore utility that works with Dockerised databases.

### Features

- **Backup**: Create compressed or uncompressed database backups
- **Restore**: Restore databases from backup files
- **List**: View all available backups with size and date
- **Clean-up**: Automatically remove old backups based on retention policy
- **Docker Integration**: Works seamlessly with Docker secrets and containers
- **Compression**: Automatic gzip compression to save storage space
- **Safety**: Confirmation prompts for destructive operations

### Quick Start

```bash
# Create a backup
./infra/scripts/db-backup.sh backup

# List all backups
./infra/scripts/db-backup.sh list

# Restore from a backup
./infra/scripts/db-backup.sh restore --file storage/backups/oullin_db_20260102_153045.sql.gz

# Clean up old backups (keeps last 7 days by default)
./infra/scripts/db-backup.sh cleanup
```

### Makefile Commands (infra/makefile/backup.mk)

You can use the Makefile shortcuts defined in `infra/makefile/backup.mk`:

```bash
# Create a backup
make backup:create

# List backups
make backup:list

# Restore from a backup
make backup:restore file=storage/backups/oullin_db_20260102_153045.sql.gz

# Cleanup old backups (default retention: 7 days)
make backup:cleanup

# Cleanup with a custom retention period
make BACKUP_RETENTION_DAYS=14 backup:cleanup

# Setup cron-based backups (weekly Sundays at 2 AM)
make backup:cron:setup

# Setup cron-based backups with a custom schedule
make backup:cron:setup schedule="0 2 * * *"

# Preview cron changes
make backup:cron:setup:dry-run

# Show or remove cron jobs
make backup:cron:show
make backup:cron:remove
```

### Commands

#### backup
Create a new database backup.

```bash
# Create a compressed backup (default)
./infra/scripts/db-backup.sh backup

# Create an uncompressed backup
./infra/scripts/db-backup.sh --compress false backup
```

Backups are stored in `storage/backups/` by default with the naming format:
`oullin_db_YYYYMMDD_HHMMSS.sql.gz`

#### restore
Restore the database from a backup file.

```bash
./infra/scripts/db-backup.sh restore --file storage/backups/oullin_db_20260102_153045.sql.gz
```

**Warning**: This will replace all data in the current database. A confirmation prompt will be shown.

#### list
Display all available backups with their size and creation date.

```bash
./infra/scripts/db-backup.sh list
```

#### cleanup
Remove backups older than the retention period (default: 7 days).

```bash
# Use default retention (7 days)
./infra/scripts/db-backup.sh cleanup

# Specify custom retention period
./infra/scripts/db-backup.sh --retention 30 cleanup
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `-h, --help` | Show help message | - |
| `-f, --file FILE` | Backup file to restore (required for restore) | - |
| `-c, --compress BOOL` | Compress backup with gzip | `true` |
| `-r, --retention DAYS` | Number of days to keep backups | `7` |
| `-d, --dir DIR` | Backup directory | `storage/backups` |

### Environment Variables

You can customise the script behaviour using environment variables:

```bash
# Custom backup directory
export BACKUP_DIR="/path/to/backups"
./infra/scripts/db-backup.sh backup

# Custom database container name
export DB_CONTAINER_NAME="my_postgres_container"
./infra/scripts/db-backup.sh backup

# Custom retention period
export BACKUP_RETENTION_DAYS=30
./infra/scripts/db-backup.sh cleanup
```

### Automated Backups with Cron

To schedule automated backups, add a cron job:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * cd /path/to/oullin/api && ./infra/scripts/db-backup.sh backup >> /var/log/db-backup.log 2>&1

# Add weekly cleanup on Sundays at 3 AM
0 3 * * 0 cd /path/to/oullin/api && ./infra/scripts/db-backup.sh cleanup >> /var/log/db-backup.log 2>&1
```

For a more robust setup with proper logging and error handling, see the example in `setup-cron-backup.sh`.

#### VPS setup (servers)

On VPS servers, use the Makefile helper so it wires up logging and cleanup for you:

```bash
# SSH into the server and move into the repo
cd /path/to/oullin/api

# Install weekly backups + cleanup (default retention: 7 days)
make backup:cron:setup

# Or install a daily backup at 2 AM with 14-day retention
make BACKUP_RETENTION_DAYS=14 backup:cron:setup schedule="0 2 * * *"

# Verify cron entries
make backup:cron:show

# Check logs
tail -f storage/logs/db-backup.log
```

The cron jobs created by `setup-cron-backup.sh` run backups and cleanup from the repo root and log to `storage/logs/db-backup.log`.

Prereqs on VPS:
- `cron` (or `cronie`) installed and enabled
- `docker` and the database container running

### Storage Locations

By default, backups are stored in:
```
storage/backups/
├── oullin_db_20260102_020000.sql.gz
├── oullin_db_20260103_020000.sql.gz
└── oullin_db_20260104_020000.sql.gz
```

You can change this location using the `--dir` option or `BACKUP_DIR` environment variable.

### Backup Format

The script uses `pg_dump` with the following options:
- `--format=plain`: SQL text format (easy to inspect and edit)
- `--no-owner`: Don't include ownership commands
- `--no-privileges`: Don't include privilege commands
- `--no-acl`: Don't include ACL commands

This makes backups portable and easy to restore on different systems.

### Security Considerations

1. **Credentials**: The script reads database credentials from Docker secrets, ensuring sensitive data isn't exposed
2. **Backups**: Store backups in a secure location with appropriate permissions
3. **Retention**: Configure appropriate retention periods to balance storage costs and recovery needs
4. **Encryption**: Consider encrypting backups for production environments
5. **Off-site**: For production, implement off-site backup storage (S3, etc.)

### Troubleshooting

#### Container not running
```
[ERROR] Database container 'oullin_db' is not running
```
Start the database container:
```bash
docker compose up -d api-db
```

#### Permission denied
```
permission denied: ./infra/scripts/db-backup.sh
```
Make the script executable:
```bash
chmod +x infra/scripts/db-backup.sh
```

#### Failed to read credentials
```
[ERROR] Failed to read database credentials from Docker secrets
```
Ensure the container is running and secrets are properly configured in `docker-compose.yml`.

### Examples

#### Daily Backup Workflow
```bash
# Create a backup
./infra/scripts/db-backup.sh backup

# Verify the backup was created
./infra/scripts/db-backup.sh list

# Test restore in a development environment
./infra/scripts/db-backup.sh restore --file storage/backups/oullin_db_20260102_153045.sql.gz
```

#### Disaster Recovery
```bash
# 1. List available backups
./infra/scripts/db-backup.sh list

# 2. Restore from the most recent backup
./infra/scripts/db-backup.sh restore --file storage/backups/oullin_db_20260104_020000.sql.gz

# 3. Verify the restoration
docker exec oullin_db psql -U <your_username> -d oullin_db -c "SELECT COUNT(*) FROM your_table;"
```

You can fetch the database username from Docker secrets inside the container:

```bash
docker exec oullin_db cat /run/secrets/pg_username
```

#### Migration to New Server
```bash
# On old server: Create backup
./infra/scripts/db-backup.sh backup

# Copy backup to new server
scp storage/backups/oullin_db_20260102_153045.sql.gz user@newserver:/path/to/oullin/api/storage/backups/

# On new server: Restore
./infra/scripts/db-backup.sh restore --file storage/backups/oullin_db_20260102_153045.sql.gz
```

### Integration with CI/CD

The backup script can be integrated into your deployment pipeline:

```yaml
# Example GitHub Actions workflow
- name: Backup database before deployment
  run: |
    ./infra/scripts/db-backup.sh backup

- name: Deploy application
  run: |
    # Your deployment commands

- name: Cleanup old backups
  run: |
    ./infra/scripts/db-backup.sh cleanup --retention 30
```
