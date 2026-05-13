# PreDeploy Guard

PreDeploy Guard is a lightweight local-first pre-deployment validation tool for backend services.

It is designed for developers, small teams, and student teams that do not have a full multi-stage deployment setup. Instead of deploying directly after building a service, developers can run PreDeploy Guard locally to create a temporary sandbox, start the service with its dependencies, run readiness, smoke, and performance checks, then generate deployment safety reports.

PreDeploy Guard also reduces YAML configuration friction through starter config generation, dependency presets, config validation, config explanation, and validation profiles.

---

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

PreDeploy Guard helps reduce this risk by providing a simple local validation layer before deployment.

---

## Solution

PreDeploy Guard creates a temporary Docker Compose sandbox for the target service.

The tool reads a `predeploy.yaml` file, builds the service image if needed, starts dependency containers, waits for dependency and service readiness, runs smoke checks, runs Dockerized k6 performance checks, captures logs on failure, and writes Markdown and JSON reports.

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
Markdown + JSON Reports
```

The long-term direction is to keep the CLI as the core engine and later add a GDP-inspired GUI/config platform on top of it.

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

### Reporting

- Markdown report generation for humans
- JSON report generation for automation and CI/CD
- Reports written next to the `predeploy.yaml` file under `reports/`
- Build, readiness, smoke, and performance sections
- Container logs included on failure

### Config usability

- `predeploy init` starter config generation
- Dependency presets through `predeploy init --with postgres,redis`
- `predeploy validate` for config validation
- `predeploy explain` for human-readable execution plans
- Validation profiles through `--profile`
- Config-relative path resolution

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
| Example Service | Go |
| Example Dependencies | PostgreSQL, Redis |
| CI Example | GitHub Actions |

---

## Repository Structure

```txt
predeploy-guard/
├── .github/
│   └── workflows/
│       └── predeploy.yml
├── cmd/
│   └── predeploy/
│       ├── main.go
│       ├── validate_output.go
│       └── explain_output.go
├── internal/
│   ├── builder/
│   ├── checker/
│   ├── config/
│   ├── loadtest/
│   ├── report/
│   ├── runner/
│   ├── sandbox/
│   └── scaffold/
├── examples/
│   ├── predeploy.yaml
│   ├── reports/
│   └── sample-service/
├── go.mod
├── go.sum
└── README.md
```

Main package boundaries:

| Package | Responsibility |
|---|---|
| `cmd/predeploy` | CLI command wiring and terminal output |
| `internal/config` | Config loading, validation, profile application, path resolution |
| `internal/scaffold` | Starter config and dependency preset generation |
| `internal/builder` | Docker image build execution |
| `internal/sandbox` | Docker Compose sandbox creation and cleanup |
| `internal/checker` | HTTP smoke/readiness checks |
| `internal/loadtest` | Dockerized k6 execution and metric parsing |
| `internal/report` | Markdown and JSON report generation |
| `internal/runner` | Main orchestration flow |

---

## Prerequisites

Make sure you have the following installed:

- Go
- Docker
- Docker Compose

Verify Docker is running:

```bash
docker version
docker compose version
```

No local k6 installation is required. PreDeploy Guard runs k6 through Docker using the `grafana/k6` image.

---

## Quick Start

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd predeploy-guard
```

### 2. Validate the example config

```bash
go run ./cmd/predeploy validate examples/predeploy.yaml
```

The validate command checks the YAML config, resolves build paths relative to the config file, and prints a summary without starting Docker.

### 3. Explain the execution plan

```bash
go run ./cmd/predeploy explain examples/predeploy.yaml
```

This explains what PreDeploy Guard will do before actually running validation.

### 4. Run PreDeploy Guard

```bash
go run ./cmd/predeploy run examples/predeploy.yaml
```

The tool will:

1. Read `examples/predeploy.yaml`
2. Build the sample backend service image
3. Create a temporary Docker Compose sandbox
4. Start PostgreSQL, Redis, and the target service
5. Wait for dependency readiness
6. Wait for service readiness
7. Run smoke checks
8. Run Dockerized k6 performance checks
9. Generate Markdown and JSON reports
10. Clean up the sandbox

---

## CLI Commands

### `predeploy init`

Creates a starter `predeploy.yaml`.

```bash
predeploy init
predeploy init --output predeploy.yaml
predeploy init --output predeploy.yaml --force
predeploy init --with postgres,redis
```

Supported dependency presets:

| Preset | Description |
|---|---|
| `postgres` / `postgresql` | Adds PostgreSQL dependency, readiness check, and `DATABASE_URL` |
| `redis` | Adds Redis dependency, readiness check, and `REDIS_URL` |

---

### `predeploy validate`

Validates the config without running Docker.

```bash
predeploy validate predeploy.yaml
predeploy validate predeploy.yaml --profile light-load
```

The output includes:

- Config file path
- Config directory
- Runtime
- Service details
- Resolved build context
- Dependencies
- Smoke checks
- Performance settings
- Available profiles

---

### `predeploy explain`

Explains what a config will do.

```bash
predeploy explain predeploy.yaml
predeploy explain predeploy.yaml --profile smoke-only
```

This is useful when a developer wants to understand the YAML without reading the full README.

---

### `predeploy run`

Runs the full validation flow.

```bash
predeploy run predeploy.yaml
predeploy run predeploy.yaml --profile light-load
```

---

## Example Configuration

`examples/predeploy.yaml`

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

    - name: list bookings
      method: GET
      path: /api/bookings
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

    - name: bookings load
      method: GET
      path: /api/bookings

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

        - name: bookings load
          method: GET
          path: /api/bookings

  stress-test:
    performance:
      enabled: true
      vus: 50
      duration: 30s
      thresholds:
        maxP95LatencyMs: 800
        maxErrorRate: 0.05
      endpoints:
        - name: health check load
          method: GET
          path: /health

        - name: bookings load
          method: GET
          path: /api/bookings

settings:
  cleanup: true
  timeoutSeconds: 60
```

Important: `build.context` is resolved relative to the `predeploy.yaml` file location.

For example, because `examples/predeploy.yaml` is inside `examples/`, this:

```yaml
build:
  context: sample-service
```

resolves to:

```txt
examples/sample-service
```

For a real service repository, the config will usually look like:

```yaml
service:
  build:
    context: .
    dockerfile: Dockerfile
```

---

## Configuration Reference

### `runtime`

```yaml
runtime:
  type: docker-compose
```

Currently supported:

| Runtime | Status |
|---|---|
| `docker-compose` | Supported |
| `kubernetes` | Planned |

---

### `service`

Defines the backend service to validate.

| Field | Description |
|---|---|
| `name` | Logical service name |
| `image` | Docker image tag for the service |
| `build.context` | Optional Docker build context, resolved relative to `predeploy.yaml` |
| `build.dockerfile` | Optional Dockerfile name or path relative to the build context |
| `port` | Container port exposed by the service |
| `healthPath` | HTTP endpoint used for service readiness |
| `env` | Environment variables passed to the service |

If `build.context` is provided, PreDeploy Guard builds the image before starting the sandbox.

---

### `dependencies`

Defines supporting services required by the target service.

| Field | Description |
|---|---|
| `image` | Docker image for the dependency |
| `port` | Internal dependency port |
| `env` | Environment variables |
| `readiness.command` | Exec-style readiness command |
| `readiness.shell` | Shell-style readiness command |
| `readiness.intervalSeconds` | Interval between readiness checks |
| `readiness.timeoutSeconds` | Maximum wait time |

Use `command` when the command exits non-zero on failure.

Use `shell` when output matching or shell pipes are needed.

Example:

```yaml
readiness:
  shell: "redis-cli ping | grep PONG"
```

---

### `checks.smoke`

Defines basic HTTP checks to run after the service becomes ready.

| Field | Description |
|---|---|
| `name` | Human-readable check name |
| `method` | HTTP method |
| `path` | Request path |
| `expectedStatus` | Expected HTTP status code |

---

### `performance`

Defines Dockerized k6 performance checks.

| Field | Description |
|---|---|
| `enabled` | Whether to run k6 performance checks |
| `vus` | Number of virtual users |
| `duration` | Duration of the k6 run |
| `thresholds.maxP95LatencyMs` | Maximum allowed p95 latency in milliseconds |
| `thresholds.maxErrorRate` | Maximum allowed HTTP error rate |
| `endpoints` | Endpoints to hit during the k6 run |

k6 is run through Docker using the `grafana/k6` image.

---

### `profiles`

Profiles allow multiple validation modes for the same service.

```yaml
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
```

Run a profile:

```bash
predeploy run predeploy.yaml --profile smoke-only
```

Profile overrides currently replace whole sections. For example, if a profile overrides `performance`, include the full performance block unless the override is intentionally minimal, such as disabling performance entirely.

---

### `settings`

Controls runtime behavior.

| Field | Description |
|---|---|
| `cleanup` | Whether to remove the sandbox after the run |
| `timeoutSeconds` | Timeout for service readiness |

---

## Reports

PreDeploy Guard generates both Markdown and JSON reports under a `reports/` directory next to the `predeploy.yaml` file.

- Markdown reports are for humans.
- JSON reports are for automation and CI/CD.

Example report sections:

```md
# PreDeploy Guard Report

## Summary

- Service: booking-service
- Image: predeploy-sample-service:latest
- Runtime: docker-compose
- Result: PASS

## Build Check

| Step | Image | Context | Dockerfile | Result | Error |
|---|---|---|---|---|---|
| docker build | predeploy-sample-service:latest | sample-service | Dockerfile | PASS | - |

## Readiness Checks

| Check | Target | Result | Error |
|---|---|---|---|
| dependency readiness | postgres | PASS | - |
| dependency readiness | redis | PASS | - |
| service readiness | http://localhost:53241/health | PASS | - |

## Smoke Checks

| Check | Method | URL | Expected | Actual | Duration | Result | Error |
|---|---|---|---:|---:|---:|---|---|
| health check | GET | http://localhost:53241/health | 200 | 200 | 15ms | PASS | - |

## Performance Check

### Latency Metrics

| Metric | Value | Threshold | Result |
|---|---:|---:|---|
| avg latency | 17.15ms | - | - |
| p95 latency | 44.29ms | 300.00ms | PASS |

### Reliability Metrics

| Metric | Value | Threshold | Result |
|---|---:|---:|---|
| error rate | 0.0000 | 0.0100 | PASS |
| request count | 300 | - | - |
```

If a failure occurs, the report also includes container logs.

---

## Failure Handling

PreDeploy Guard captures useful debugging information when a run fails.

Failure scenarios include:

- Docker image build failure
- Dependency readiness failure
- Service readiness failure
- Smoke check failure
- k6 performance threshold failure
- k6 execution failure

For example, if Redis readiness fails, the report may show:

```md
## Readiness Checks

| Check | Target | Result | Error |
|---|---|---|---|
| dependency readiness | redis | FAIL | container predeploy-redis became unhealthy |
| service readiness | http://localhost:53241/health | FAIL | service readiness skipped because dependency readiness failed |
```

The report will also include container logs to help diagnose the issue.

---

## GitHub Actions

PreDeploy Guard can run in GitHub Actions as a CI validation gate.

Example workflow:

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

The workflow fails if PreDeploy Guard fails, and reports are uploaded as artifacts for debugging.

---

## Design Notes

### Why Docker Compose first?

The initial goal is to support local pre-deployment validation for small developers and teams.

Docker Compose is simpler than Kubernetes for this use case because it avoids requiring:

- Cluster creation
- Kubeconfig setup
- Namespace management
- Ingress configuration
- Kubernetes networking debugging

Kubernetes support can be added later as a separate runtime.

---

### Why generate a temporary sandbox?

PreDeploy Guard avoids modifying the developer’s existing environment.

Each run creates a temporary Compose environment, runs checks, generates a report, and optionally cleans up everything after completion.

---

### Why dependency readiness matters

Starting a dependency container does not mean the dependency is ready.

For example:

- PostgreSQL may need time to initialize
- Redis may start quickly but still needs to respond correctly
- A backend service may fail if it starts before dependencies are ready

PreDeploy Guard checks dependency readiness before validating the target service.

---

### Why config usability commands matter

A major goal is to reduce YAML configuration mistakes.

The CLI now includes:

```bash
predeploy init
predeploy validate
predeploy explain
```

These commands are early steps toward a future GUI/config wizard.

---

## Current Limitations

This is still a local-first validation tool. Current limitations include:

- Docker Compose runtime only
- Kubernetes runtime not implemented yet
- No GUI/config wizard yet
- No API gateway or ingress latency simulation yet
- Dependency readiness depends on user-provided commands
- Smoke checks currently support simple HTTP status validation
- Profile overrides are section-level replacements rather than deep merges
- Secrets are passed as plain environment variables in local configs

---

## Roadmap

### Completed

- Phase 1: Docker Compose validation MVP
  - Docker Compose sandbox
  - image build
  - dependency readiness
  - service readiness
  - smoke checks
  - Markdown reports

- Phase 2: Performance and automation
  - Dockerized k6
  - JSON reports
  - richer performance metrics
  - config-relative path resolution
  - GitHub Actions workflow support

- Phase 3: Config usability
  - `predeploy init`
  - dependency presets
  - `predeploy validate`
  - `predeploy explain`
  - validation profiles

### Next

- Phase 4: Run history and profile comparison
  - store previous run summaries
  - compare latest run against previous run
  - compare profiles such as `smoke-only` vs `light-load`
  - CLI command for listing and viewing run history

### Future

- GDP-inspired GUI/platform mode
  - guided YAML builder
  - dependency wizard
  - smoke check builder
  - k6 profile builder
  - report dashboard
  - GitHub Actions workflow generator

- Kubernetes runtime
  - existing cluster support
  - temporary namespace creation
  - manifest-based deployment
  - Helm chart support

- API gateway and ingress testing
  - gateway configuration
  - routing latency testing
  - direct service vs gateway latency comparison

---

## Development Commands

Run the tool:

```bash
go run ./cmd/predeploy run examples/predeploy.yaml
```

Validate config:

```bash
go run ./cmd/predeploy validate examples/predeploy.yaml
```

Explain config:

```bash
go run ./cmd/predeploy explain examples/predeploy.yaml
```

Run a profile:

```bash
go run ./cmd/predeploy run examples/predeploy.yaml --profile smoke-only
go run ./cmd/predeploy run examples/predeploy.yaml --profile light-load
```

Generate a starter config:

```bash
go run ./cmd/predeploy init --with postgres,redis --output predeploy.yaml
```

Format code:

```bash
go fmt ./...
```

Run basic checks:

```bash
go vet ./...
```

Build CLI binary:

```bash
go build -o predeploy ./cmd/predeploy
```

Run with binary:

```bash
./predeploy run examples/predeploy.yaml
```

On Windows PowerShell:

```powershell
.\predeploy.exe run examples/predeploy.yaml
```

---

## Project Positioning

PreDeploy Guard is best described as:

> A local-first pre-deployment validation tool that creates temporary Docker Compose sandboxes for backend services, validates dependency readiness, service readiness, smoke checks, and performance thresholds, then generates human-readable and machine-readable deployment safety reports.

This project demonstrates backend engineering, developer tooling, Docker-based workflows, performance validation, CI readiness, reliability thinking, and early platform engineering concepts.
