### Ollin's API

:sparkles: Every time you visit my site, behind the scenes, an API written in Go is hard at work delivering content,
handling requests, and powering dynamic features.

:monorail: The **oullin/api** repository is that engine. It bundles all the application’s core logic, data access, and
configuration into a clean, maintainable service, making it the indispensable backbone of the “Ollin” experience.

:hearts: In short, **oullin/api** isn’t just another code repository—it’s the beating heart of my web application. It translates
every user action into data operations and returns precisely what the frontend needs.

:rocket: Feel free to explore the folders, clone the repository and run it locally via Docker Compose. If you feel adventurous,
consider contributing to the project by making improvements or fixing issues by sending a pull request.

> This is where the mindful movement of "Ollin" truly comes alive, one request at a time.

---

### Updating Grafana Dashboards Safely

To keep dashboard changes reproducible and under version control:

1. **Start monitoring stack**: `make monitor-up`
2. **Make changes in Grafana UI**: Navigate to http://localhost:3000 and edit dashboards
3. **Export your changes**: Run `./infra/metrics/grafana/scripts/export-dashboards.sh`
   - Select specific dashboard or `all` to export all dashboards
   - Exports are saved to `infra/metrics/grafana/dashboards/`
4. **Review the diff**: `git diff infra/metrics/grafana/dashboards/`
5. **Commit changes**: Add and commit the exported JSON files
6. **Verify**: `make monitor-restart` to ensure dashboards reload correctly

:warning: **Always export after making UI changes**—manual edits to JSON files can work but are error-prone.

---

### Metrics Endpoint Security

The `/metrics` endpoint uses **network isolation**, not authentication (commit `3b5d07e`).

**Security Model:**
- Port `9180` uses `expose:` in `docker-compose.yml` (NOT `ports:`)—only accessible via Docker internal network
- Caddyfile serves `/metrics` on `:9180` server block (internal only)
- Public domains (`oullin.io`, etc.) have **no** `/metrics` routes

**Regression Prevention:**
- Never publish port `9180` to the host (no `ports: - "9180:9180"`)
- Never add `/metrics` handlers to public-facing Caddy server blocks
- Network isolation is the industry standard (Google, Netflix, Uber)

:lock: **Do not revert to auth-based metrics**—Prometheus cannot generate dynamic signatures for scraping.
