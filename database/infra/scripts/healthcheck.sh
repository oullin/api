#!/bin/sh

# Exit immediately if any command fails.
set -e

# Read the secrets into variables. This is more robust than direct command substitution.
DB_USER=$(cat /run/secrets/pg_username)
DB_NAME=$(cat /run/secrets/pg_dbname)

# Explicitly check if the user variable is empty. If it is, fail immediately.
# This prevents the "role -d does not exist" error.
if [ -z "$DB_USER" ]; then
  echo "Healthcheck Error: The pg_username secret is empty or could not be read." >&2
  exit 1
fi

# Execute the final command. 'exec' replaces the shell process, which is slightly more efficient.
# The variables are double-quoted to handle any special characters safely.
exec pg_isready -U "$DB_USER" -d "$DB_NAME"
