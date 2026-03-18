### 🚀 Ollin's API

<details>
<summary><strong>About the Project</strong></summary>

:sparkles: Every time you visit my site, behind the scenes, an API written in Go is hard at work delivering content,
handling requests, and powering dynamic features.

:monorail: The **oullin/api** repository is that engine. It bundles all the application’s core logic, data access, and
configuration into a clean, maintainable service, making it the indispensable backbone of the “Ollin” experience.

:hearts: In short, **oullin/api** isn’t just another code repository—it’s the beating heart of my web application. It translates
every user action into data operations and returns precisely what the frontend needs.

:rocket: Feel free to explore the folders, clone the repository and run it locally via Docker Compose. If you feel adventurous,
consider contributing to the project by making improvements or fixing issues by sending a pull request.

> This is where the mindful movement of “Ollin” truly comes alive, one request at a time.

</details>

### ⚡ Quick Start

1.  **Configure**: `cp .env.example .env`
2.  **Run in background**: `make build-local` (Builds and runs the local stack detached)
3.  **Run with attached logs**: `make watch-local` (Starts the Docker local stack in the foreground)

For more details, check the [Setup Guide](docs/SETUP.md).

### 🗄️ Documentation

**Get Started**

- [Setup & Development](docs/SETUP.md)
- [API Reference](docs/API.md)
- [Database backups](docs/DB_BACKUPS.md)


**Infrastructure & Ops**

- [Metrics & Monitoring](infra/metrics/README.md)
    - [Grafana Diagnostics](infra/metrics/docs/GRAFANA_DIAGNOSTICS.md)
    - [VPS Deployment](infra/metrics/docs/VPS_DEPLOYMENT.md)
- [Caddy Server](infra/caddy/readme.md)
- [Database SSL](database/infra/ssl/readme.md)
