# PreDeploy Guard
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![k6](https://img.shields.io/badge/k6-Performance_Testing-7D64FF?logo=k6&logoColor=white)
![React](https://img.shields.io/badge/React-Dashboard-61DAFB?logo=react&logoColor=black)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-CI-2088FF?logo=githubactions&logoColor=white)

> Local-first pre-deployment validation and experiment sandbox for backend services.

PreDeploy Guard helps backend developers catch deployment issues before release by spinning up temporary Docker Compose sandboxes, validating dependencies, running smoke checks, executing k6 performance tests, and generating Markdown/JSON safety reports. Its next direction is config-driven local deployment experiments for testing services, infrastructure components, and workloads before real deployment.

It is designed for developers, small teams, and student teams that do not have a full multi-stage deployment setup. Instead of deploying directly after building a service, developers can run PreDeploy Guard locally to create a temporary sandbox, start the service with its dependencies, run readiness, smoke, and performance checks, then generate deployment safety reports.

PreDeploy Guard also reduces YAML configuration friction through starter config generation, dependency presets, config validation, config explanation, validation profiles, run history, a local API layer, and a local dashboard UI.

---

## Demo

### Dashboard

![PreDeploy Guard Dashboard](assets/dashboard.png)

### Validation Report

![Validation Report](assets/report.png)

## Problem

Small teams and solo developers often do not have proper staging environments.

This means a backend service may be deployed without knowing whether:

- The container image builds correctly
- Required dependencies such as PostgreSQL or Redis are reachable
- The service can start successfully
- The health endpoint is working
- Basic API endpoints return expected status codes
- Performance checks pass latency and error-rate thresholds
- Logs contain obvious runtime errors
- The validation configuration itself is correct
- Recent changes caused a regression compared to a previous run

PreDeploy Guard helps reduce this risk by providing a simple local validation layer before deployment.

---

## Solution

PreDeploy Guard creates a temporary Docker Compose sandbox for the target service.

The tool reads a `predeploy.yaml` file, builds the service image if needed, starts dependency containers, waits for dependency and service readiness, runs smoke checks, runs Dockerized k6 performance checks, captures logs on failure, writes Markdown and JSON reports, and records run history.

```txt
Developer Service Repo
        |
        | predeploy.yaml
        v
PreDeploy Guard CLI
        |
        | builds image if configured
        v
Temporary Docker Compose Sandbox
        |
        | starts service + dependencies
        v
Readiness + Smoke + Performance Checks
        |
        v
Markdown + JSON Reports + Run History
```

PreDeploy Guard also includes a local API server and dashboard:

```txt
Local Dashboard UI
        |
        | HTTP API
        v
PreDeploy Guard Local Server
        |
        | shared app services
        v
Config Summary + Run History + Reports + API-triggered Runs
```

The long-term direction is to keep the CLI as the core engine and evolve the local sandbox toward richer deployment experiments. Docker Compose remains the supported default runtime; future Kubernetes support will target existing kubeconfig contexts, including local clusters such as Minikube, kind, and k3s.

---

## Project Scope

PreDeploy Guard is best described as:

> A local-first pre-deployment validation and experiment sandbox that creates isolated environments for backend services, validates readiness, smoke checks, and performance thresholds, and produces human-readable and machine-readable reports.

PreDeploy Guard is intended for local testing and experimentation, not production deployment orchestration.

This project demonstrates backend engineering, developer tooling, Docker-based workflows, performance validation, CI readiness, reliability thinking, local API design, frontend dashboard development, and early platform engineering concepts.

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

### Local API and dashboard

- `predeploy serve` local API server
- Safe config summary and explanation endpoints
- History and report API endpoints
- API-triggered validation runs
- Async run task tracking
- Task status polling
- Basic task logs endpoint
- React + TypeScript dashboard under `web/dashboard`
- Dashboard cards for API health and config summary
- Run history table
- Run details panel
- Markdown report preview
- Run task table
- Task logs panel
- Trigger run panel with profile selector

### CI support

- GitHub Actions workflow support
- Reports uploaded as workflow artifacts
- CI fails when PreDeploy Guard validation fails

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
| Dashboard | React, TypeScript, Vite |
| CI Example | GitHub Actions |

---

## What This Project Demonstrates

PreDeploy Guard demonstrates:

- Backend engineering with Go and clean package boundaries
- REST API design for local validation runs, report access, and task polling
- Docker Compose-based dependency orchestration
- Performance validation using Dockerized k6
- Developer tooling and platform engineering thinking
- CI/CD readiness through GitHub Actions integration
- Local dashboard development with React, TypeScript, and Vite
- Reliability-focused design through readiness checks, smoke checks, run history, and report comparison

---

## Repository Structure

```txt
predeploy-guard/
├── .github/
│   └── workflows/
│       └── predeploy.yml
├── cmd/
│   └── predeploy/
├── internal/
│   ├── app/
│   ├── builder/
│   ├── checker/
│   ├── config/
│   ├── history/
│   ├── loadtest/
│   ├── report/
│   ├── runner/
│   ├── sandbox/
│   ├── scaffold/
│   └── server/
├── web/
│   └── dashboard/
│       ├── index.html
│       ├── package.json
│       ├── vite.config.ts
│       ├── tsconfig.json
│       └── src/
│           ├── api/
│           ├── components/
│           ├── types/
│           ├── utils/
│           ├── App.tsx
│           ├── main.tsx
│           └── styles.css
├── examples/
│   ├── predeploy.yaml
│   └── sample-service/
├── AGENTS.md
├── ROADMAP.md
├── go.mod
├── go.sum
└── README.md
```

Main package boundaries:

| Package | Responsibility |
|---|---|
| `cmd/predeploy` | CLI command wiring and terminal output |
| `internal/app` | Shared use-case services for CLI and server |
| `internal/config` | Config loading, validation, profile application, path resolution |
| `internal/scaffold` | Starter config and dependency preset generation |
| `internal/builder` | Docker image build execution |
| `internal/sandbox` | Docker Compose sandbox creation and cleanup |
| `internal/checker` | HTTP smoke/readiness checks |
| `internal/loadtest` | Dockerized k6 execution and metric parsing |
| `internal/report` | Markdown and JSON report generation |
| `internal/history` | Run history index, lookup, and comparison helpers |
| `internal/server` | Local HTTP API server for dashboard support |
| `internal/runner` | Main validation orchestration flow |
| `web/dashboard` | Local React dashboard |

---

## Prerequisites

Backend:

- Go
- Docker
- Docker Compose

Dashboard:

- Node.js
- npm

Verify Docker is running:

```bash
docker version
docker compose version
```

No local k6 installation is required. PreDeploy Guard runs k6 through Docker using the `grafana/k6` image.

---

## Quick Start

### CLI validation

```bash
git clone https://github.com/HXong/predeploy-guard.git
cd predeploy-guard

go run ./cmd/predeploy validate examples/predeploy.yaml
go run ./cmd/predeploy explain examples/predeploy.yaml
go run ./cmd/predeploy run examples/predeploy.yaml
```

### Local API server

```bash
go run ./cmd/predeploy serve --config examples/predeploy.yaml
```

Default address:

```txt
http://localhost:7070
```

### Dashboard

In another terminal:

```bash
cd web/dashboard
npm install
npm run dev
```

Open:

```txt
http://localhost:5173
```

During development, the Vite dev server proxies `/api` requests to `http://localhost:7070`.

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

## Dashboard

The dashboard lives under:

```txt
web/dashboard
```

Current dashboard capabilities:

- View API health
- View config summary
- View execution plan summary
- Trigger validation run by profile
- View async run tasks
- Poll selected task status
- View task logs
- View run history
- Select a run
- View run details
- Open Markdown/JSON report links
- Preview raw Markdown report content
- Switch between the dashboard and guided Config Builder

### Guided Config Builder

The dashboard Config Builder provides forms for service configuration, dependencies and readiness, HTTP smoke checks, k6 performance checks, and named validation profiles. The generated `predeploy.yaml` preview updates as the configuration changes and can be copied or downloaded for local use.

The builder exports YAML in the browser. It does not write configuration files through the local API.

Frontend structure:

```txt
web/dashboard/src/
├── app/
│   ├── App.tsx
│   └── AppShell.tsx
├── features/
│   ├── config/
│   ├── config-builder/
│   ├── dashboard/
│   └── runs/
├── shared/
│   ├── api/
│   ├── components/
│   └── utils/
├── main.tsx
└── styles.css
```

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

## Current Limitations

- Docker Compose runtime only
- Kubernetes runtime not implemented yet
- Guided config builder previews and exports YAML in the browser; it does not write configuration files through the backend
- No API gateway or ingress latency simulation yet
- Dependency readiness depends on user-provided commands
- Smoke checks currently support simple HTTP status validation
- Profile overrides are section-level replacements rather than deep merges
- Secrets are passed as plain environment variables in local configs
- Run comparison currently compares summary-level metrics only
- API-triggered run logs currently contain high-level task events only, not full runner step output
- Dashboard previews raw Markdown only; it does not render Markdown as HTML yet

---

## Roadmap

### Completed

- Phase 1: Docker Compose validation MVP
- Phase 2: Performance and automation
- Phase 3: Config usability
- Phase 4: Run history and comparison
- Phase 5: Local API and GUI groundwork
- Phase 6: Local dashboard UI
- Phase 7: Guided config builder

### Next

- Phase 8: Deployment Experiment Sandbox
  - align runtime and experiment models while preserving Docker Compose
  - prepare runtime adapter boundaries
  - add future kubeconfig-based local-cluster execution, generic experiment workloads, and experiment reporting incrementally

### Future

- Optional safe config save-to-file workflow
- GitHub Actions workflow generator
- Helm chart support
- API gateway and ingress testing

---

## Development Commands

Backend:

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

Frontend:

```bash
cd web/dashboard
npm install
npm run dev
npm run build
```


