# Setup & Development Guide

## Prerequisites

- **Go**: Version 1.22+ (Check `go.mod` for exact version)
- **Docker**: For running the database and monitoring stack.
- **Make**: For running project commands.
- No manual Docker volume setup is required. The Make targets create the external DB volume automatically when needed.

## Configuration

Copy the example environment file to `.env`:

```bash
cp .env.example .env
```

Review the `.env` file and adjust the settings as needed.

### Key Environment Variables

- `ENV_APP_NAME`: Name of the application.
- `ENV_APP_ENV_TYPE`: Environment type (e.g., `local`, `production`).
- `ENV_APP_LOGS_DIR`: Application log filename pattern inside the API runtime.
- `API_LOGS_PATH`: Host path mounted to `/app/storage/logs` for persistent API logs.
- `ENV_DB_*`: Database connection details.
- `ENV_HTTP_PORT`: Port for the HTTP server (default: `8080`).

## Running the Application

### Quick Start

To start the application with a fresh state:

```bash
make fresh
```

This will clean logs and build artifacts.

### Database

The application uses a PostgreSQL database. You can manage it using the following commands:

- Start DB: `make db:up`
- Run Migrations: `make db:migrate`
- Seed Data: `make db:seed`
- Reset DB (Fresh): `make db:fresh`
- Check Connection: `make db:ping`

### Building

- Build for local development: `make build-local`
- Build for release: `make build-release`

### Running

To run the application locally:

- **Optional first-run prewarm**: `make prewarm-cli-docker` warms the Docker CLI module cache, build cache, and reusable CLI binary. It does **not** start the database.
- **CLI Mode**: `make run-cli` reuses `oullin_db` when it is already healthy, starts `api-db` only when needed, and reuses a Docker-built CLI binary on warm runs.
- **Metal (Dev) Mode**: `make run-metal`

### Monitoring

The project includes a monitoring stack with Prometheus and Grafana.

- Start Monitoring: `make monitor-up`
- Stop Monitoring: `make monitor-down`
- Check Status: `make monitor-status`
- Open Grafana: `make monitor-grafana`

For production outage investigation, use [Uptime Incident Log Runbook](UPTIME_LOGS.md).

## Testing

Run the test suite:

```bash
make test-all
```

## Code Quality

- Format code: `make format`
- Run audit/checks: `make audit`
