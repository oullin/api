#!/bin/sh
set -e

# Read Docker secrets and construct DATA_SOURCE_NAME
export DATA_SOURCE_NAME="postgresql://$(cat /run/secrets/pg_username):$(cat /run/secrets/pg_password)@api-db:5432/$(cat /run/secrets/pg_dbname)?sslmode=require"

# Execute postgres_exporter with any additional arguments
exec /bin/postgres_exporter "$@"
