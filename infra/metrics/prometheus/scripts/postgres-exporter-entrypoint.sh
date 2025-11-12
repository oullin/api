#!/bin/sh
set -e

# URL-encode function using od and tr (POSIX-compliant)
# Required for credentials containing special characters (@, :, /, ?, =)
urlencode() {
  string="$1"
  printf '%s' "$string" | od -An -tx1 | tr ' ' % | tr -d '\n'
}

# Read Docker secrets separately for better error diagnostics
PG_USER=$(cat /run/secrets/pg_username)
PG_PASSWORD=$(cat /run/secrets/pg_password)
PG_DBNAME=$(cat /run/secrets/pg_dbname)

# Construct DATA_SOURCE_NAME with URL-encoded credentials
export DATA_SOURCE_NAME="postgresql://$(urlencode "$PG_USER"):$(urlencode "$PG_PASSWORD")@api-db:5432/$(urlencode "$PG_DBNAME")?sslmode=require"

# Execute postgres_exporter with any additional arguments
exec /bin/postgres_exporter "$@"
