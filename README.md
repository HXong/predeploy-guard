# PreDeploy Guard

PreDeploy Guard is a lightweight local-first pre-deployment validation tool for backend services.

It is designed for developers, small teams, and student teams that do not have a full multi-stage deployment setup. Instead of deploying directly after building a service, developers can run PreDeploy Guard locally to create a temporary sandbox, start the service with its dependencies, run readiness, smoke, and performance checks, then generate deployment safety reports.

PreDeploy Guard also reduces YAML configuration friction through starter config generation, dependency presets, config validation, config explanation, validation profiles, run history, and a local API layer for future GUI integration.

---

## Current Features

### Core validation

- CLI-based workflow
- YAML-based configuration
- Docker image build support
- Temporary Docker Compose sandbox
- Automatic free host port allocation
- Dependency service support
- Dependency readiness checks
- Service readiness checks
- HTTP smoke checks
- Dockerized k6 performance checks
- Performance thresholds for p95 latency and error rate
- Rich performance metrics
- Container log collection on failure
- Automatic cleanup

### Reporting and history

- Markdown report generation for humans
- JSON report generation for automation and CI/CD
- Reports written next to the `predeploy.yaml` file under `reports/`
- `reports/history.json` run history index
- `predeploy history` to list previous runs
- `predeploy show` to inspect a previous run
- `predeploy compare` to compare two previous runs

### Config usability

- `predeploy init` starter config generation
- Dependency presets through `predeploy init --with postgres,redis`
- `predeploy validate` for config validation
- `predeploy explain` for human-readable execution plans
- Validation profiles through `--profile`
- Config-relative path resolution

### Local API / GUI groundwork

- `predeploy serve` local API server
- Safe config summary and explanation endpoints
- History and report API endpoints
- API-triggered validation runs
- Async run task tracking
- Task status polling
- Basic task logs endpoint
- Server route tests

---

## Tech Stack

| Area | Technology |
|---|---|
| CLI / Core Tool | Go |
| Config | YAML |
| Sandbox Runtime | Docker Compose |
| Service Build | Docker |
| Smoke Checks | Go `net/http` |
| Performance Checks | Dockerized k6 |
| Reports | Markdown, JSON |
| History | JSON index |
| Local API | Go `net/http` |
| CI Example | GitHub Actions |

---

## Quick Start

```bash
git clone <your-repo-url>
cd predeploy-guard
go run ./cmd/predeploy validate examples/predeploy.yaml
go run ./cmd/predeploy explain examples/predeploy.yaml
go run ./cmd/predeploy run examples/predeploy.yaml
```

The run command will build the target image if configured, create a temporary Docker Compose sandbox, start dependencies and the service, run readiness/smoke/performance checks, write reports, append history, and clean up.

---

## CLI Commands

### `predeploy init`

```bash
predeploy init
predeploy init --output predeploy.yaml
predeploy init --output predeploy.yaml --force
predeploy init --with postgres,redis
```

### `predeploy validate`

```bash
predeploy validate predeploy.yaml
predeploy validate predeploy.yaml --profile light-load
```

### `predeploy explain`

```bash
predeploy explain predeploy.yaml
predeploy explain predeploy.yaml --profile smoke-only
```

### `predeploy run`

```bash
predeploy run predeploy.yaml
predeploy run predeploy.yaml --profile light-load
```

### `predeploy history`

```bash
predeploy history predeploy.yaml
```

### `predeploy show`

```bash
predeploy show predeploy.yaml <run-id>
```

### `predeploy compare`

```bash
predeploy compare predeploy.yaml <base-run-id> <target-run-id>
```

### `predeploy serve`

```bash
predeploy serve --config examples/predeploy.yaml
```

Default address:

```txt
http://localhost:7070
```

---

## Local API Endpoints

### Health

```txt
GET /api/health
```

### Config

```txt
GET /api/config/summary
GET /api/config/explain
```

These endpoints return safe, GUI-friendly config views. They do not expose the raw internal config struct.

### History and reports

```txt
GET /api/history
GET /api/history/{runId}
GET /api/reports/{runId}/json
GET /api/reports/{runId}/markdown
```

### API-triggered runs

```txt
POST /api/runs
GET  /api/runs
GET  /api/runs/{taskId}
GET  /api/runs/{taskId}/logs
```

PowerShell example:

```powershell
Invoke-RestMethod -Uri "http://localhost:7070/api/runs" -Method POST -ContentType "application/json" -Body '{"profile":"smoke-only"}'
```

Git Bash / Linux / macOS example:

```bash
curl -X POST "http://localhost:7070/api/runs" \
  -H "Content-Type: application/json" \
  -d '{"profile":"smoke-only"}'
```

`POST /api/runs` returns immediately with a task ID. The task status can then be polled through `GET /api/runs/{taskId}`.

---

## Example Configuration

```yaml
runtime:
  type: docker-compose

service:
  name: booking-service
  image: predeploy-sample-service:latest
  build:
    context: sample-service
    dockerfile: Dockerfile
  port: 8080
  healthPath: /health
  env:
    PORT: "8080"
    DATABASE_URL: postgres://test:test@postgres:5432/testdb
    REDIS_URL: redis://redis:6379

dependencies:
  postgres:
    image: postgres:16
    port: 5432
    env:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: testdb
    readiness:
      command: ["pg_isready", "-U", "test", "-d", "testdb"]
      intervalSeconds: 2
      timeoutSeconds: 30

  redis:
    image: redis:7
    port: 6379
    readiness:
      shell: "redis-cli ping | grep PONG"
      intervalSeconds: 2
      timeoutSeconds: 30

checks:
  smoke:
    - name: health check
      method: GET
      path: /health
      expectedStatus: 200

performance:
  enabled: true
  vus: 10
  duration: 15s
  thresholds:
    maxP95LatencyMs: 300
    maxErrorRate: 0.01
  endpoints:
    - name: health check load
      method: GET
      path: /health

profiles:
  smoke-only:
    performance:
      enabled: false

  light-load:
    performance:
      enabled: true
      vus: 10
      duration: 15s
      thresholds:
        maxP95LatencyMs: 300
        maxErrorRate: 0.01
      endpoints:
        - name: health check load
          method: GET
          path: /health

settings:
  cleanup: true
  timeoutSeconds: 60
```

`build.context` is resolved relative to the `predeploy.yaml` file location.

---

## Reports and History

PreDeploy Guard generates both Markdown and JSON reports under a `reports/` directory next to the `predeploy.yaml` file.

It also maintains:

```txt
reports/history.json
```

Markdown reports are for humans. JSON reports and `history.json` are for automation, dashboards, and future GUI support.

---

## GitHub Actions

PreDeploy Guard can run in GitHub Actions as a CI validation gate.

```yaml
name: PreDeploy Guard

on:
  push:
    branches:
      - main
  pull_request:

jobs:
  predeploy:
    name: Run PreDeploy Guard
    runs-on: ubuntu-latest

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Verify Docker
        run: |
          docker version
          docker compose version

      - name: Download Go dependencies
        run: go mod download

      - name: Validate PreDeploy config
        run: go run ./cmd/predeploy validate examples/predeploy.yaml

      - name: Run PreDeploy Guard
        run: go run ./cmd/predeploy run examples/predeploy.yaml

      - name: Upload PreDeploy reports
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: predeploy-reports
          path: examples/reports/
```

---

## Current Limitations

- Docker Compose runtime only
- Kubernetes runtime not implemented yet
- No GUI/config wizard yet
- No API gateway or ingress latency simulation yet
- Dependency readiness depends on user-provided commands
- Smoke checks currently support simple HTTP status validation
- Profile overrides are section-level replacements rather than deep merges
- Secrets are passed as plain environment variables in local configs
- Run comparison currently compares summary-level metrics only
- API-triggered run logs currently contain high-level task events only, not full runner step output

---

## Roadmap

### Completed

- Phase 1: Docker Compose validation MVP
- Phase 2: Performance and automation
- Phase 3: Config usability
- Phase 4: Run history and comparison
- Phase 5: Local API and GUI groundwork

### Next

- Phase 6: Local dashboard UI
  - React + TypeScript frontend
  - API health/config summary cards
  - run history table
  - run task table
  - trigger validation run by profile
  - task status polling
  - basic task logs panel

### Future

- GDP-inspired GUI/platform mode
- Guided YAML/config builder
- Dependency wizard
- Smoke check builder
- k6 profile builder
- GitHub Actions workflow generator
- Kubernetes runtime
- Helm chart support
- API gateway and ingress testing

---

## Development Commands

```bash
go run ./cmd/predeploy validate examples/predeploy.yaml
go run ./cmd/predeploy explain examples/predeploy.yaml
go run ./cmd/predeploy run examples/predeploy.yaml
go run ./cmd/predeploy run examples/predeploy.yaml --profile smoke-only
go run ./cmd/predeploy history examples/predeploy.yaml
go run ./cmd/predeploy show examples/predeploy.yaml <run-id>
go run ./cmd/predeploy compare examples/predeploy.yaml <base-run-id> <target-run-id>
go run ./cmd/predeploy serve --config examples/predeploy.yaml
go fmt ./...
go vet ./...
go test ./...
```

---

## Project Positioning

PreDeploy Guard is best described as:

> A local-first pre-deployment validation tool that creates temporary Docker Compose sandboxes for backend services, validates dependency readiness, service readiness, smoke checks, and performance thresholds, then generates human-readable and machine-readable deployment safety reports with run history, comparison, and local API support.

This project demonstrates backend engineering, developer tooling, Docker-based workflows, performance validation, CI readiness, reliability thinking, local API design, and early platform engineering concepts.
