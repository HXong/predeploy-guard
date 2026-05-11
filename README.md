# PreDeploy Guard

PreDeploy Guard is a lightweight local pre-deployment validation tool for backend services.

It is designed for developers or small teams that do not have a full multi-stage deployment setup. Instead of deploying directly after building a service, developers can run PreDeploy Guard locally to create a temporary sandbox, start the service with its dependencies, run readiness and smoke checks, and generate a deployment safety report.

---

## Problem

Small teams, student teams, and solo developers often do not have proper staging environments.

This means a backend service may be deployed without knowing whether:

- The container image builds correctly
- Required dependencies such as PostgreSQL or Redis are reachable
- The service can start successfully
- The health endpoint is working
- Basic API endpoints return the expected status codes
- Logs contain obvious runtime errors

PreDeploy Guard helps reduce this risk by providing a simple local validation layer before deployment.

---

## Solution

PreDeploy Guard creates a temporary Docker Compose sandbox for the target service.

The tool reads a `predeploy.yaml` file, builds the service image if needed, starts dependency containers, waits for readiness checks, runs smoke tests, captures logs on failure, and writes a Markdown report.

```txt
Developer Service
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
Readiness Checks + Smoke Checks
        |
        v
Markdown Deployment Report
```

---

## Features

Current features:

- CLI-based workflow
- YAML configuration
- Docker image build support
- Temporary Docker Compose sandbox
- Automatic free host port allocation
- Dependency service support
- Dependency readiness checks
- Service readiness checks
- HTTP smoke checks
- Container log collection on failure
- Markdown and JSON report generation
- Automatic cleanup
- Example Go backend service with PostgreSQL and Redis
- Dockerized k6 performance checks
- Performance thresholds for p95 latency and error rate
- Rich performance metrics
- Markdown and JSON report generation
- Config-relative path resolution
- Config validation command
- GitHub Actions workflow support

---

## Tech Stack

| Area | Technology |
|---|---|
| CLI / Core Tool | Go |
| Config | YAML |
| Sandbox Runtime | Docker Compose |
| Service Build | Docker |
| Smoke Checks | Go `net/http` |
| Reports | Markdown |
| Example Service | Go |
| Example Dependencies | PostgreSQL, Redis |

---

## Repository Structure

```txt
predeploy-guard/
├── cmd/
│   └── predeploy/
│       └── main.go
├── internal/
│   ├── builder/
│   │   └── docker.go
│   ├── checker/
│   │   ├── smoke.go
│   │   └── url.go
│   ├── config/
│   │   ├── config.go
│   │   └── validation.go
│   ├── report/
│   │   ├── report.go
│   │   └── markdown.go
│   ├── runner/
│   │   ├── runner.go
│   │   ├── output.go
│   │   └── logs.go
│   └── sandbox/
│       ├── compose.go
│       ├── compose_file.go
│       ├── file.go
│       ├── health.go
│       ├── naming.go
│       └── port.go
├── examples/
│   ├── predeploy.yaml
│   └── sample-service/
│       ├── Dockerfile
│       ├── go.mod
│       ├── go.sum
│       └── main.go
├── reports/
├── go.mod
├── go.sum
└── README.md
```

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

---

## Quick Start

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd predeploy-guard
```

### 2. PreDeploy validate
```bash
go run ./cmd/predeploy validate examples/predeploy.yaml
```

The validate command checks the YAML config, resolves build paths relative to the config file, and prints a summary without starting Docker.

### 3. Run PreDeploy Guard

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
8. Run k6 performance checks
9. Generate Markdown and JSON reports
10. Clean up the sandbox

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

settings:
  cleanup: true
  timeoutSeconds: 60
```

---

## Configuration Reference

### `runtime`

Defines the execution runtime.

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

```yaml
service:
  name: booking-service
  image: predeploy-sample-service:latest
  build:
    context: examples/sample-service
    dockerfile: Dockerfile
  port: 8080
  healthPath: /health
  env:
    PORT: "8080"
```

| Field | Description |
|---|---|
| `name` | Logical service name |
| `image` | Docker image tag for the service |
| `build.context` | Optional Docker build context |
| `build.dockerfile` | Optional Dockerfile name |
| `port` | Container port exposed by the service |
| `healthPath` | HTTP endpoint used for readiness |
| `env` | Environment variables passed to the service |

If `build.context` is provided, PreDeploy Guard builds the image before starting the sandbox.

---

### `dependencies`

Defines supporting services required by the target service.

```yaml
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
```

Each dependency can include:

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

```yaml
checks:
  smoke:
    - name: health check
      method: GET
      path: /health
      expectedStatus: 200
```

| Field | Description |
|---|---|
| `name` | Human-readable check name |
| `method` | HTTP method |
| `path` | Request path |
| `expectedStatus` | Expected HTTP status code |

---

### `settings`

Controls runtime behavior.

```yaml
settings:
  cleanup: true
  timeoutSeconds: 60
```

| Field | Description |
|---|---|
| `cleanup` | Whether to remove the sandbox after the run |
| `timeoutSeconds` | Timeout for service readiness |

---

## Example Output

Successful run:

```txt
Loaded config successfully
Service: booking-service
Image: predeploy-sample-service:latest
Port: 8080
Health path: /health
Checking whether target image build is required...
Built target image successfully: predeploy-sample-service:latest
Creating Docker Compose sandbox...
Sandbox directory: /tmp/predeploy-guard-xxxxxx
Starting service with Docker Compose...
Service base URL: http://localhost:53241
Waiting for dependency readiness...
Dependency ready: postgres
Dependency ready: redis
Waiting for service readiness...
Running smoke checks...
[PASS] health check GET http://localhost:53241/health expected=200 actual=200
[PASS] list bookings GET http://localhost:53241/api/bookings expected=200 actual=200
Result: PASS
Report written to: reports/predeploy-booking-service-2026-05-08-120000.md
Stopping Docker Compose sandbox...
Removing sandbox files...
```

---

## Report Example

PreDeploy Guard generates both Markdown and JSON reports under a `reports/` directory next to the `predeploy.yaml` file.

- Markdown reports are for humans.
- JSON reports are for automation and CI/CD.

Example sections:

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
| docker build | predeploy-sample-service:latest | examples/sample-service | Dockerfile | PASS | - |

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

## Current Limitations

This is the Phase 2 MVP, so there are several known limitations:

- Docker Compose runtime only
- Local development focused
- Kubernetes runtime not implemented yet
- No GUI/config wizard yet
- No API gateway or ingress latency simulation yet
- Dependency readiness depends on user-provided commands
- Smoke checks currently support simple HTTP status validation

---

## Roadmap

### Completed:
- Phase 1: Docker Compose validation MVP
- Phase 2: Dockerized k6, JSON reports, richer metrics, validation command, GitHub Actions workflow

### Next Developements:
- Phase 3: Config usability
  - predeploy init
  - config templates
  - dependency presets
  - profile support

### Future:
- GUI/platform mode inspired by GDP
  - guided YAML builder
  - dependency wizard
  - k6 profile builder
  - report dashboard
  - GitHub Actions workflow generator
  - Kubernetes runtime support

---

## Development Commands

Run the tool:

```bash
go run ./cmd/predeploy run examples/predeploy.yaml
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

> A local pre-deployment validation tool that creates temporary Docker Compose sandboxes for backend services, validates dependency readiness, service readiness, and smoke checks, then generates deployment safety reports.

This project demonstrates backend engineering, developer tooling, Docker-based workflows, reliability thinking, and early platform engineering concepts.
