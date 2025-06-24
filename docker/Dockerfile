# Filename: Dockerfile.go
# Use Docker BuildKit syntax for advanced features.
#syntax=docker/dockerfile:1.4

# --- Build Stage ---
# Use the official Go image, specifically version 1.24, as the builder.
# Using alpine for a smaller base image.
FROM golang:1.24-alpine AS builder

# Install necessary build tools. 'git' is required for Go modules,
# 'openssh-client' for SSH, and 'tzdata' for timezone info.
RUN apk add --no-cache git tzdata openssh-client

# Set the working directory inside the container.
WORKDIR /app

# Before fetching dependencies, configure git to use SSH for GitHub URLs.
# This forces git to use SSH instead of HTTPS.
RUN git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"

# Add GitHub's public key to the known_hosts file to prevent
# "Host key verification failed" errors.
RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Declare a build argument for the SSH private key.
# This key will be passed in during the build process.
ARG SSH_PRIVATE_KEY

# Write the provided SSH private key to the standard key location
# and set the correct file permissions (600). This is critical for SSH security.
RUN echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa && \
    chmod 600 ~/.ssh/id_rsa

# Copy the Go module files from the build context (project root).
COPY go.mod go.sum ./

# Copy the application source code.
COPY main.go ./

# Run go mod tidy to sync dependencies. It will now use the id_rsa file for auth.
RUN go mod tidy

# Download dependencies listed in go.mod.
RUN go mod download

# Build the Go application.
# -o /app/server creates the binary named 'server' in the /app directory.
# CGO_ENABLED=0 disables Cgo, which is needed for a static binary.
# -ldflags="-s -w" strips debugging information, reducing the binary size.
RUN CGO_ENABLED=0 go build -o /app/server -ldflags="-s -w" .

# --- Final Stage ---
# Use a minimal, non-root base image for the final container for security.
FROM alpine:latest

# We'll create a non-root user for security purposes.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory.
WORKDIR /home/appuser

# Copy only the compiled binary from the builder stage.
COPY --from=builder /app/server .

# Copy timezone data from the builder stage.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Ensure the new user owns the application files.
RUN chown -R appuser:appgroup /home/appuser

# Switch to the non-root user.
USER appuser

# Expose the port the app runs on.
EXPOSE 8080

# The command to run when the container starts.
CMD ["./server"]
