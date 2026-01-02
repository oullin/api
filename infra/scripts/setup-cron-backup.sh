#!/usr/bin/env bash
set -euo pipefail

# Setup script for automated database backups via cron
# This script helps configure cron jobs for regular database backups

SCRIPT_DIR=""
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR

PROJECT_ROOT=""
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly PROJECT_ROOT

readonly LOG_DIR="${PROJECT_ROOT}/storage/logs"
readonly BACKUP_SCRIPT="${SCRIPT_DIR}/db-backup.sh"

# Colors
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly RED='\033[0;31m'
readonly NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

require_arg() {
    if [[ -z "${2-}" || "${2-}" == -* ]]; then
        log_error "Missing or invalid value for ${1}"
        show_usage
        exit 1
    fi
}

show_usage() {
    cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Setup automated database backups using cron.

OPTIONS:
    -h, --help              Show this help message
    -s, --schedule CRON     Cron schedule (default: "0 2 * * 0" - weekly on Sundays at 2 AM)
    -r, --retention DAYS    Backup retention in days (default: 7)
    --cleanup-schedule CRON Cleanup schedule (default: "0 3 * * 0" - Sundays at 3 AM)
    --dry-run               Show what would be done without making changes

EXAMPLES:
    # Setup with defaults (weekly backup on Sundays at 2 AM, cleanup on Sundays at 3 AM)
    $(basename "$0")

    # Custom schedule (daily backups at 2 AM)
    $(basename "$0") --schedule "0 2 * * *"

    # Preview changes without applying
    $(basename "$0") --dry-run

COMMON CRON SCHEDULES:
    "0 2 * * 0"      Weekly on Sunday at 2:00 AM
    "0 2 * * *"      Daily at 2:00 AM
    "0 */6 * * *"    Every 6 hours
    "0 3 1 * *"      Monthly on the 1st at 3:00 AM
    "*/30 * * * *"   Every 30 minutes

EOF
}

check_prerequisites() {
    if [[ ! -f "${BACKUP_SCRIPT}" ]]; then
        log_error "Backup script not found: ${BACKUP_SCRIPT}"
        exit 1
    fi

    if [[ ! -x "${BACKUP_SCRIPT}" ]]; then
        log_error "Backup script is not executable: ${BACKUP_SCRIPT}"
        log_info "Run: chmod +x ${BACKUP_SCRIPT}"
        exit 1
    fi

    if ! command -v crontab &> /dev/null; then
        log_error "crontab command not found. Cron may not be installed."
        exit 1
    fi
}

create_log_directory() {
    if [[ ! -d "${LOG_DIR}" ]]; then
        log_info "Creating log directory: ${LOG_DIR}"
        mkdir -p "${LOG_DIR}"
    fi
}

generate_cron_entries() {
    local backup_schedule="$1"
    local cleanup_schedule="$2"
    local retention_days="$3"
    local log_file="${LOG_DIR}/db-backup.log"

    cat <<EOF
# Oullin API Database Backup - Auto-generated
# DO NOT EDIT MANUALLY - Use setup-cron-backup.sh to modify

# Automated database backup
${backup_schedule} cd "${PROJECT_ROOT}" && "${BACKUP_SCRIPT}" backup >> "${log_file}" 2>&1

# Periodic cleanup of old backups (retention: ${retention_days} days)
${cleanup_schedule} cd "${PROJECT_ROOT}" && BACKUP_RETENTION_DAYS=${retention_days} "${BACKUP_SCRIPT}" cleanup >> "${log_file}" 2>&1

EOF
}

install_cron_jobs() {
    local backup_schedule="$1"
    local cleanup_schedule="$2"
    local retention_days="$3"
    local dry_run="${4:-false}"

    log_info "Generating cron entries..."
    local new_entries
    new_entries=$(generate_cron_entries "${backup_schedule}" "${cleanup_schedule}" "${retention_days}")

    if [[ "${dry_run}" == "true" ]]; then
        log_warn "DRY RUN - The following would be added to crontab:"
        echo
        echo "${new_entries}"
        echo
        log_info "Run without --dry-run to apply changes"
        return 0
    fi

    # Get current crontab, remove old Oullin backup entries, add new ones
    local temp_cron
    temp_cron=$(mktemp)
    trap 'rm -f "${temp_cron}"' RETURN

    # Export current crontab (ignore error if empty)
    crontab -l 2>/dev/null | grep -Fv "Oullin API Database Backup" | grep -Fv "${BACKUP_SCRIPT}" > "${temp_cron}" || true

    # Add new entries
    echo "${new_entries}" >> "${temp_cron}"

    # Install new crontab
    if crontab "${temp_cron}"; then
        log_info "Cron jobs installed successfully"
    else
        log_error "Failed to install cron jobs"
        exit 1
    fi

    log_info "Backup schedule: ${backup_schedule}"
    log_info "Cleanup schedule: ${cleanup_schedule}"
    log_info "Retention period: ${retention_days} days"
    log_info "Logs: ${LOG_DIR}/db-backup.log"
}

show_current_crontab() {
    log_info "Current crontab entries for database backups:"
    echo
    if crontab -l 2>/dev/null | grep -A 5 -F "Oullin API Database Backup"; then
        echo
    else
        log_warn "No existing backup cron jobs found"
    fi
}

test_backup() {
    log_info "Testing backup script..."

    if "${BACKUP_SCRIPT}" --help &>/dev/null; then
        log_info "Backup script is working correctly"
    else
        log_error "Backup script test failed"
        exit 1
    fi
}

remove_cron_jobs() {
    log_info "Removing Oullin database backup cron jobs..."

    local temp_cron
    temp_cron=$(mktemp)
    trap 'rm -f "${temp_cron}"' RETURN

    # Export current crontab and remove Oullin backup entries
    if crontab -l 2>/dev/null | grep -Fv "Oullin API Database Backup" | grep -Fv "${BACKUP_SCRIPT}" > "${temp_cron}"; then
        crontab "${temp_cron}"
        log_info "Cron jobs removed successfully"
    else
        log_warn "No cron jobs found to remove"
    fi

}

main() {
    local backup_schedule="0 2 * * 0"  # Weekly on Sundays at 2 AM
    local cleanup_schedule="0 3 * * 0" # Sundays at 3 AM
    local retention_days=7
    local dry_run=false
    local action="install"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -s|--schedule)
                require_arg "$1" "${2-}"
                backup_schedule="$2"
                shift 2
                ;;
            -r|--retention)
                require_arg "$1" "${2-}"
                retention_days="$2"
                shift 2
                ;;
            --cleanup-schedule)
                require_arg "$1" "${2-}"
                cleanup_schedule="$2"
                shift 2
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --remove)
                action="remove"
                shift
                ;;
            --show)
                action="show"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Execute action
    case "${action}" in
        install)
            check_prerequisites
            create_log_directory
            test_backup
            install_cron_jobs "${backup_schedule}" "${cleanup_schedule}" "${retention_days}" "${dry_run}"

            if [[ "${dry_run}" != "true" ]]; then
                echo
                log_info "Setup complete! Your database will be backed up automatically."
                log_info "To view logs: tail -f ${LOG_DIR}/db-backup.log"
                log_info "To verify cron jobs: crontab -l | grep -A 5 'Oullin API Database Backup'"
            fi
            ;;
        remove)
            remove_cron_jobs
            ;;
        show)
            show_current_crontab
            ;;
    esac
}

main "$@"
