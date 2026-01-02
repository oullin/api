#!/usr/bin/env bash
set -euo pipefail

# Database Backup Script for PostgreSQL
# Supports backup, restore, and automatic cleanup operations

# --- Configuration
SCRIPT_DIR=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

PROJECT_ROOT=""
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly PROJECT_ROOT

readonly BACKUP_DIR="${BACKUP_DIR:-${PROJECT_ROOT}/storage/backups}"
readonly CONTAINER_NAME="${DB_CONTAINER_NAME:-oullin_db}"
readonly RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"

TIMESTAMP=""
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
readonly TIMESTAMP

# --- Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# --- Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

show_usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS] COMMAND

Database backup and restore utility for Oullin API.

COMMANDS:
    backup      Create a new database backup
    restore     Restore database from a backup file
    list        List all available backups
    cleanup     Remove backups older than retention period

OPTIONS:
    -h, --help              Show this help message
    -f, --file FILE         Backup file to restore (required for restore)
    -c, --compress          Compress backup with gzip (default: enabled)
    -r, --retention DAYS    Number of days to keep backups (default: ${RETENTION_DAYS})
    -d, --dir DIR           Backup directory (default: ${BACKUP_DIR})

EXAMPLES:
    # Create a backup
    $(basename "$0") backup

    # Create an uncompressed backup
    $(basename "$0") --compress=false backup

    # List all backups
    $(basename "$0") list

    # Restore from a specific backup
    $(basename "$0") restore --file storage/backups/oullin_db_20260102_153045.sql.gz

    # Clean up old backups (keeps last ${RETENTION_DAYS} days)
    $(basename "$0") cleanup

ENVIRONMENT VARIABLES:
    BACKUP_DIR              Override default backup directory
    DB_CONTAINER_NAME       Override database container name (default: oullin_db)
    BACKUP_RETENTION_DAYS   Override retention period in days (default: ${RETENTION_DAYS})

EOF
}

check_container() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        log_error "Database container '${CONTAINER_NAME}' is not running"
        log_info "Start it with: docker compose up -d api-db"
        exit 1
    fi
}

get_db_credentials() {
    # Read credentials from Docker secrets inside the container
    DB_USER=""
    DB_USER=$(docker exec "${CONTAINER_NAME}" cat /run/secrets/pg_username 2>/dev/null || echo "")
    DB_NAME=""
    DB_NAME=$(docker exec "${CONTAINER_NAME}" cat /run/secrets/pg_dbname 2>/dev/null || echo "")

    if [[ -z "${DB_USER}" || -z "${DB_NAME}" ]]; then
        log_error "Failed to read database credentials from Docker secrets"
        exit 1
    fi
}

create_backup() {
    local compress="${1:-true}"

    log_info "Starting database backup..."

    check_container
    get_db_credentials

    # Create backup directory if it doesn't exist
    mkdir -p "${BACKUP_DIR}"

    local backup_file="${BACKUP_DIR}/oullin_db_${TIMESTAMP}.sql"

    log_info "Database: ${DB_NAME}"
    log_info "User: ${DB_USER}"
    log_info "Backup file: ${backup_file}"

    # Create the backup using pg_dump inside the container
    if docker exec "${CONTAINER_NAME}" pg_dump \
        -U "${DB_USER}" \
        -d "${DB_NAME}" \
        --verbose \
        --format=plain \
        --no-owner \
        --no-privileges \
        --no-acl > "${backup_file}"; then

        log_info "Backup created successfully"

        # Compress if requested
        if [[ "${compress}" == "true" ]]; then
            log_info "Compressing backup..."
            if gzip -f "${backup_file}"; then
                backup_file="${backup_file}.gz"
                log_info "Backup compressed: ${backup_file}"
            else
                log_warn "Compression failed, keeping uncompressed backup"
            fi
        fi

        # Show backup size
        local size
        size=$(du -h "${backup_file}" | cut -f1)
        log_info "Backup size: ${size}"
        log_info "Backup completed successfully: ${backup_file}"

        return 0
    else
        log_error "Backup failed"
        rm -f "${backup_file}"
        exit 1
    fi
}

restore_backup() {
    local backup_file="$1"

    if [[ -z "${backup_file}" ]]; then
        log_error "Backup file is required for restore"
        show_usage
        exit 1
    fi

    # Handle relative paths
    if [[ ! "${backup_file}" =~ ^/ ]]; then
        backup_file="${PROJECT_ROOT}/${backup_file}"
    fi

    if [[ ! -f "${backup_file}" ]]; then
        log_error "Backup file not found: ${backup_file}"
        exit 1
    fi

    check_container
    get_db_credentials

    log_warn "This will REPLACE all data in database '${DB_NAME}'"
    read -p "Are you sure you want to continue? (yes/no): " -r
    echo

    if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
        log_info "Restore cancelled"
        exit 0
    fi

    log_info "Starting database restore..."
    log_info "Backup file: ${backup_file}"

    # Decompress if needed and pipe to container
    if [[ "${backup_file}" =~ \.gz$ ]]; then
        log_info "Decompressing and restoring backup..."
        if gunzip -c "${backup_file}" | docker exec -i "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}"; then
            log_info "Restore completed successfully"
        else
            log_error "Restore failed"
            exit 1
        fi
    else
        log_info "Restoring backup..."
        if docker exec -i "${CONTAINER_NAME}" psql -U "${DB_USER}" -d "${DB_NAME}" < "${backup_file}"; then
            log_info "Restore completed successfully"
        else
            log_error "Restore failed"
            exit 1
        fi
    fi
}

list_backups() {
    if [[ ! -d "${BACKUP_DIR}" ]]; then
        log_warn "Backup directory does not exist: ${BACKUP_DIR}"
        return 0
    fi

    log_info "Available backups in ${BACKUP_DIR}:"
    echo

    # Find all .sql and .sql.gz files, sort by modification time
    local backups
    backups=$(find "${BACKUP_DIR}" -maxdepth 1 -type f \( -name "*.sql" -o -name "*.sql.gz" \) -print0 | xargs -0 ls -lt 2>/dev/null)

    if [[ -z "${backups}" ]]; then
        log_warn "No backups found"
        return 0
    fi

    printf "%-50s %-10s %-20s\n" "FILE" "SIZE" "DATE"
    printf "%-50s %-10s %-20s\n" "----" "----" "----"

    echo "${backups}" | while read -r line; do
        local file
        file=$(echo "$line" | awk '{print $NF}')
        local size
        size=$(echo "$line" | awk '{print $5}')
        local date
        date=$(echo "$line" | awk '{print $6, $7, $8}')
        local filename
        filename=$(basename "$file")

        # Convert size to human readable
        local human_size
        human_size=$(du -h "$file" | cut -f1)

        printf "%-50s %-10s %-20s\n" "$filename" "$human_size" "$date"
    done

    echo
    local total
    total=$(find "${BACKUP_DIR}" -maxdepth 1 -type f \( -name "*.sql" -o -name "*.sql.gz" \) | wc -l | tr -d ' ')
    log_info "Total backups: ${total}"
}

cleanup_old_backups() {
    if [[ ! -d "${BACKUP_DIR}" ]]; then
        log_warn "Backup directory does not exist: ${BACKUP_DIR}"
        return 0
    fi

    log_info "Cleaning up backups older than ${RETENTION_DAYS} days..."

    local count=0
    while IFS= read -r -d '' file; do
        log_info "Removing: $(basename "$file")"
        rm -f "$file"
        ((count++))
    done < <(find "${BACKUP_DIR}" -maxdepth 1 -type f \( -name "*.sql" -o -name "*.sql.gz" \) -mtime "+${RETENTION_DAYS}" -print0)

    if [[ $count -eq 0 ]]; then
        log_info "No old backups to remove"
    else
        log_info "Removed ${count} old backup(s)"
    fi
}

# --- Main script
main() {
    local command=""
    local backup_file=""
    local compress="true"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -f|--file)
                backup_file="$2"
                shift 2
                ;;
            -c|--compress)
                compress="$2"
                shift 2
                ;;
            -r|--retention)
                RETENTION_DAYS="$2"
                shift 2
                ;;
            -d|--dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            backup|restore|list|cleanup)
                command="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Validate command
    if [[ -z "${command}" ]]; then
        log_error "No command specified"
        show_usage
        exit 1
    fi

    # Execute command
    case "${command}" in
        backup)
            create_backup "${compress}"
            ;;
        restore)
            restore_backup "${backup_file}"
            ;;
        list)
            list_backups
            ;;
        cleanup)
            cleanup_old_backups
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
