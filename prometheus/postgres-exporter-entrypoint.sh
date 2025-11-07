#!/bin/sh
set -e

# Read Docker secrets separately for better error diagnostics
PG_USER=$(cat /run/secrets/pg_username)
PG_PASSWORD=$(cat /run/secrets/pg_password)
PG_DBNAME=$(cat /run/secrets/pg_dbname)

# Construct DATA_SOURCE_NAME from individual variables
export DATA_SOURCE_NAME="postgresql://${PG_USER}:${PG_PASSWORD}@api-db:5432/${PG_DBNAME}?sslmode=require"

# Execute postgres_exporter with any additional arguments
exec /bin/postgres_exporter "$@"
