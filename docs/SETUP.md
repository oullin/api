# Setup & Development Guide

## Prerequisites

- **Go**: Version 1.22+ (Check `go.mod` for exact version)
- **Docker**: For running the database and monitoring stack.
- **Make**: For running project commands.

## Configuration

Copy the example environment file to `.env`:

```bash
cp .env.example .env
```

Review the `.env` file and adjust the settings as needed.

### Key Environment Variables

- `ENV_APP_NAME`: Name of the application.
- `ENV_APP_ENV_TYPE`: Environment type (e.g., `local`, `production`).
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

- **CLI Mode**: `make run-cli`
- **Metal (Dev) Mode**: `make run-metal`

### Monitoring

The project includes a monitoring stack with Prometheus and Grafana.

- Start Monitoring: `make monitor-up`
- Stop Monitoring: `make monitor-down`
- Check Status: `make monitor-status`
- Open Grafana: `make monitor-grafana`

## Testing

Run the test suite:

```bash
make test-all
```

## Code Quality

- Format code: `make format`
- Run audit/checks: `make audit`
