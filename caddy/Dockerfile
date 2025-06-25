# Filename: caddy/Dockerfile
# This Dockerfile builds a Caddy image using a specific, stable version number.

# Define a build argument for the Caddy version with a sensible default.
# This allows the version to be easily overridden from the docker-compose.yml file.
ARG CADDY_VERSION=2.10.0

# Use the official Caddy image with the latest tag.
FROM caddy:${CADDY_VERSION}

# Copy your custom Caddyfile into the container.
# This overwrites the default Caddyfile.
COPY Caddyfile.local /etc/caddy/Caddyfile
