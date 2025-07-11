#!/bin/sh

# Exit immediately if a command exits with a non-zero status.
set -e

# Read the credentials from the Docker Secret files
DB_USER=$(cat /run/secrets/pg_username)
DB_PASSWORD=$(cat /run/secrets/pg_password)
DB_NAME=$(cat /run/secrets/pg_dbname)

# Construct the database URL using the values from the secrets
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@api-db:5432/${DB_NAME}?sslmode=require"

# Execute the migrate tool, passing the constructed URL and any other arguments
# The "$@" passes along any arguments from the docker-compose command (like "up", "down 1", etc.)
migrate -path /migrations -database "${DATABASE_URL}" "$@"
