#!/bin/sh

set -eux

# Take ownership of the main data directory.
chown -R postgres:postgres /var/lib/postgresql/data

# Decide what to run:
# 	• If the user passed arguments (via "command:"), use those.
#	• Otherwise default to our config_file flag.
if [ "$#" -gt 0 ]; then
  ARGS="$*"
else
  ARGS="postgres -c config_file=/etc/postgresql/postgresql.conf"
fi

# Fix SSL file ownership & perms
chown postgres:postgres /etc/ssl/private/server.key
chown postgres:postgres /etc/ssl/certs/server.crt
chmod 600 /etc/ssl/private/server.key
chmod 644 /etc/ssl/certs/server.crt

# Hand off to the *official* entrypoint as the postgres user
# 	- `su postgres -c "docker-entrypoint.sh $ARGS"` will:
#		• drop privileges to postgres
#		• invoke the real /usr/local/bin/docker-entrypoint.sh with our ARGS
exec su postgres -c "docker-entrypoint.sh $ARGS"
