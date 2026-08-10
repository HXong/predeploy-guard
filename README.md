# PreDeploy Guard
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![k6](https://img.shields.io/badge/k6-Performance_Testing-7D64FF?logo=k6&logoColor=white)
![React](https://img.shields.io/badge/React-Dashboard-61DAFB?logo=react&logoColor=black)
![GitHub Actions](https://img.shields.io/badge/GitHub_Actions-CI-2088FF?logo=githubactions&logoColor=white)

> Config-driven local deployment experiment sandbox for backend services.

PreDeploy Guard helps developers:

- validate service readiness;
- check dependencies;
- run smoke checks;
- test gateway and direct-service behavior;
- run workload experiments;
- run performance checks;
- review structured Markdown and JSON reports; and
- inspect results through a local dashboard.

Docker Compose is the default runtime. An early Kubernetes local-cluster runtime works through existing kubeconfig contexts and developer-managed clusters. PreDeploy Guard is a local sandbox and experiment runner, not a production deployment platform.

---

## First-time setup

Phase 10 is complete and provides the developer environment doctor, safe app-aware init, opt-in guided init, and actionable onboarding recommendations.

For an existing backend application, use this onboarding flow:

```bash
predeploy doctor --app ./my-app
predeploy init --interactive --app ./my-app
predeploy doctor --config predeploy.yaml --app ./my-app
predeploy validate predeploy.yaml
predeploy run predeploy.yaml
```

Doctor reports local readiness and prints actionable recommendations for the next command. Recommendations can guide initialization, validation, runtime checks, kubeconfig troubleshooting, ingress responsibility, or Dockerized k6 prerequisites, but they never execute commands, install tools, or include configured environment values.

For an nginx or static demo app, use the endpoint the image actually serves. The default nginx image serves `/` on port `80`, so initialize it with:

```bash
predeploy init --interactive --app ./demo-app --port 80 --health-path /
```

Use `/health` only when the application actually exposes that endpoint.

App-aware and guided init write only the selected configuration outside the application directory. PreDeploy Guard does not install Docker, kubectl, Minikube, ingress controllers, or any other tool. It does not modify app folders, create Dockerfiles, or read or print `.env` values. Guided prompts are opt-in through `--interactive` and never run by default.

---

## Dashboard preview

### Configuration overview

![PreDeploy Guard Dashboard](assets/dashboard.png)

### Structured run review

![Structured run timeline in the dashboard](assets/report.png)

Both screenshots use generic sample data and omit raw logs, secrets, and local filesystem paths. See the [demo guide](docs/demo.md) for the capture checklist.

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

## Architecture

PreDeploy Guard keeps configuration, runtime orchestration, validation, and presentation separated behind runtime-neutral boundaries:

```txt
predeploy.yaml
    |
    v
CLI / Local API
    |
    v
Runtime Adapter
    |
    v
Local Sandbox
    |
    v
Checks + Workloads + Performance
    |
    v
Markdown/JSON Reports + Run History
    |
    v
React Dashboard
```

Docker Compose remains the supported default. The Kubernetes local-cluster MVP uses `kubectl` and existing kubeconfig contexts; it does not install, start, or manage a cluster for the user.

---

## Why this project matters

PreDeploy Guard demonstrates backend and platform engineering concepts in a local-first workflow:

- runtime sandbox orchestration;
- readiness, dependency, and smoke validation;
- gateway and direct-service comparison;
- workload and performance experiments;
- structured reporting and run history;
- a local API and React dashboard; and
- explicit safety boundaries for developer tooling.

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
- Kubernetes local-cluster sandbox through existing kubeconfig contexts and `kubectl`
- Temporary Kubernetes namespace, rollout readiness, port-forwarding, diagnostics, and owned cleanup
- Automatic free host port allocation
- Dependency service support
- Dependency readiness checks
- Service readiness checks
- Optional external gateway route checks with direct-service status comparison
- Optional owned Kubernetes Ingress generation for configured gateway routes
- Optional direct-vs-gateway latency thresholds with warn or fail policies
- HTTP smoke checks
- Runtime-neutral HTTP experiment workloads with configurable rate, duration, expected status, and failure policy
- Dockerized k6 performance checks
- Performance thresholds for p95 latency and error rate
- Rich performance metrics
- Container log collection on failure
- Automatic cleanup

### Reporting and history

- Markdown report generation for humans
- JSON report generation for automation and CI/CD
- Runtime environment summaries for Docker Compose and Kubernetes sandboxes
- Run phase timelines with passed, failed, and skipped phases plus durations
- Runtime diagnostics for failed or incomplete local experiments
- Reports written next to the `predeploy.yaml` file under `reports/`
- `reports/history.json` run history index
- `predeploy history` to list previous runs
- `predeploy show` to inspect a previous run
- `predeploy compare` to compare two previous runs

### Config usability

- `predeploy doctor` for safe local environment and app-path readiness checks
- `predeploy init` starter config generation
- App-aware config generation through `predeploy init --app <folder>`
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
- Structured JSON report review for runtime environment, run timeline, gateway results, workload results, and performance results
- Markdown and JSON report links
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
| Sandbox Runtime | Docker Compose; Kubernetes local-cluster MVP via `kubectl` |
| Service Build | Docker |
| Smoke Checks | Go `net/http` |
| Gateway Checks | Go `net/http` against a user-provided external URL |
| Experiment Workloads | Go `net/http` HTTP traffic MVP |
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
| `internal/runtime` | Runtime lifecycle interfaces and Docker Compose/Kubernetes adapter selection |
| `internal/checker` | HTTP smoke/readiness checks |
| `internal/workload` | Runtime-neutral experiment workload execution |
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

Kubernetes local runtime (optional):

- `kubectl`
- An existing working cluster and kubeconfig context
- Service and dependency images accessible to that cluster

PreDeploy Guard does not install or start a cluster, and it does not automatically load locally built images into Minikube, kind, or other local clusters.

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

## Quick Demo

### 1. Validate and run the sample service

```bash
git clone https://github.com/HXong/predeploy-guard.git
cd predeploy-guard

predeploy validate examples/predeploy.yaml
predeploy run examples/predeploy.yaml
```

If the binary is not installed, use:

```bash
go run ./cmd/predeploy validate examples/predeploy.yaml
go run ./cmd/predeploy run examples/predeploy.yaml
```

Docker must be running for the sample validation. Generated reports are written under `examples/reports/` next to the config and are ignored by git.

### 2. Start the local API server

```bash
predeploy serve --config examples/predeploy.yaml --addr localhost:7070
```

Without an installed binary:

```bash
go run ./cmd/predeploy serve --config examples/predeploy.yaml --addr localhost:7070
```

### 3. Start the dashboard

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

The Vite development server proxies `/api` requests to the local API server on `localhost:7070`. For a short walkthrough of what to click and how to clean up, see [docs/demo.md](docs/demo.md).

---

## CLI Commands

### `predeploy doctor`

```bash
predeploy doctor
predeploy doctor --config predeploy.yaml
predeploy doctor --app ./my-app
predeploy doctor --config predeploy.yaml --app ./my-app
```

Doctor reports `PASS`, `WARN`, and `FAIL` results for local filesystem access, Git repository detection, configuration validity, Docker, Docker Compose, kubectl, Kubernetes connectivity, optional Minikube, and common application project files. A valid config makes checks for its selected runtime, generated ingress, and Dockerized performance engine required.

Doctor is diagnostic only. It does not install or start tools, modify application source, edit host files, run k6, or install an ingress controller. Warnings keep a zero exit code; failed required checks return exit code 1.

After its checks, doctor recommends the safest next command based on whether a config or app path was supplied and which configured runtime features need attention. Recommendations are guidance only and are never executed automatically.

### `predeploy init`

```bash
predeploy init
predeploy init --output predeploy.yaml
predeploy init --output predeploy.yaml --force
predeploy init --with postgres,redis
predeploy init --app ./my-app --port 8080 --health-path /health
predeploy init --app ./my-app --runtime kubernetes
predeploy init --app ./my-app --service-name orders-api --image orders-api:local
predeploy init --app ./my-app --no-build
predeploy init --interactive
predeploy init --interactive --app ./my-app
predeploy init --interactive --output predeploy.yaml
```

App-aware init detects only common top-level project files and links the application through `service.build.context` when a Dockerfile exists. The build context is relative to the generated config where possible. PreDeploy Guard does not modify the application folder, create a Dockerfile, parse `.env`, or overwrite an existing config unless `--force` is supplied.

For the default nginx image or another static demo that serves `/` on port `80`, use `predeploy init --interactive --app ./demo-app --port 80 --health-path /`. The `/health` default is appropriate only when the application provides that route.

Recommended onboarding flow:

```bash
predeploy doctor --app ./my-app
predeploy init --interactive --app ./my-app
predeploy doctor --config predeploy.yaml --app ./my-app
predeploy validate predeploy.yaml
predeploy run predeploy.yaml
```

When no Dockerfile is found—or when `--no-build` is used—the generated config references the selected or default image without a build context.

Guided init is strictly opt-in through `--interactive`. Existing flags pre-fill the corresponding prompts, invalid runtime, port, health-path, and dependency answers are re-prompted, and a final confirmation is required before the config is written. The wizard generates only the selected config file; it does not modify the application folder, create a Dockerfile, inspect `.env` values, or install tools.

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
- View structured runtime environment and run timeline details
- Review gateway, direct-service, latency, workload, and performance results
- Open Markdown and JSON report links
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

workloads:
  - name: warmup-traffic
    type: http
    enabled: true
    method: GET
    path: /health
    duration: 10s
    ratePerSecond: 5
    expectedStatus: 200
    failurePolicy: fail

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

## Gateway Checks

Phase 9A validates selected routes through an external gateway URL after the direct service becomes ready. Phase 9B can also generate an owned Kubernetes Ingress for those routes:

```yaml
runtime:
  type: kubernetes

gateway:
  enabled: true
  baseURL: http://predeploy.local
  ingress:
    enabled: true
    host: predeploy.local
    className: local-controller
    pathType: Prefix
  latency:
    enabled: true
    maxGatewayLatencyMs: 500
    maxOverheadMs: 200
    maxOverheadRatio: 3.0
    failurePolicy: warn
  routes:
    - name: homepage-via-gateway
      method: GET
      path: /
      expectedStatus: 200
      compareDirect: true
```

Each route is requested through `gateway.baseURL`. When `compareDirect` is enabled, the same method and path are also requested through the runtime-provided direct service URL, and both responses must have the expected, matching status. `method`, `expectedStatus`, and `compareDirect` default to `GET`, `200`, and `true`.

Ingress generation is optional, Kubernetes-only, namespace-scoped, and disabled by default. PreDeploy Guard uses the configured routes, host, class name, path type, and annotations without adding controller-specific settings.

PreDeploy Guard does not install or manage an ingress controller. `gateway.baseURL` must already resolve to the local ingress endpoint; PreDeploy Guard does not discover ingress IPs or edit host files.

Because ingress controllers synchronize new routes asynchronously, generated-ingress gateway checks retry every second until all routes pass. The retry window uses `settings.timeoutSeconds`, capped at 30 seconds. External gateway checks without generated Ingress remain single-attempt checks.

Phase 9C can evaluate the final request durations against an absolute gateway latency limit and, when direct comparison is available, maximum gateway overhead in milliseconds or as a ratio. At least one threshold is required when latency comparison is enabled. `failurePolicy` defaults to `warn`; warnings are reported without failing an otherwise correct route, while `fail` makes a threshold violation fail the route and run.

This is a lightweight comparison of the final correctness-check requests, not a load test or statistically significant benchmark. Use the existing k6 performance checks for sustained traffic and percentile-based performance validation. Generated-ingress retry decisions use HTTP correctness only, so a latency-only violation does not cause repeated readiness attempts.

Reusable examples are available under `examples/`:

- `kubernetes-ingress-gateway.yaml` demonstrates generated Kubernetes Ingress, external gateway checks, and warning-only latency thresholds. It requires an existing ingress controller and a `gateway.baseURL` that already resolves to its endpoint.
- `kubernetes-local-image.yaml` demonstrates a Kubernetes service using a locally prepared image; `local-nginx-image/Dockerfile` provides the small sample image source.

---

## Reports and History

PreDeploy Guard generates both Markdown and JSON reports under a `reports/` directory next to the `predeploy.yaml` file.

It also maintains:

```txt
reports/history.json
```

Markdown reports are for humans. JSON reports and `history.json` are for automation, dashboards, and future GUI support.

Generated reports are intentionally ignored by git. They may contain local paths, raw tool output, or environment-specific diagnostics. Commit sanitized screenshots or documentation instead of generated report artifacts.

---

## Current Limitations

- Docker Compose remains the default runtime
- Kubernetes support is an early local-cluster MVP that requires an existing working cluster, kubeconfig, and `kubectl`
- Kubernetes image loading is not implemented; configured service and dependency images must already be accessible to the selected cluster
- HTTP is the first implemented experiment workload; file/log replay, message producers/consumers, background jobs, custom commands, and Kubernetes Jobs remain future work
- Guided config builder previews and exports YAML in the browser; it does not write configuration files through the backend
- Gateway validation requires a separately managed, reachable external gateway URL
- Generated Ingress resources do not support TLS, authentication, custom headers, or Gateway API yet
- Gateway latency comparison uses one final correctness-check sample per route; it is not a substitute for k6 performance testing
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
- Phase 8: Deployment Experiment Sandbox
  - completed runtime and experiment model alignment
  - completed the runtime adapter foundation
  - completed kubeconfig-based local-cluster execution through `kubectl`
  - completed runtime-neutral HTTP experiment workloads
  - completed runtime environment, phase timeline, and diagnostics report enhancements
  - future: additional generic workload types
- Phase 9A: External gateway checks and direct-vs-gateway status comparison
- Phase 9B: Owned Kubernetes Ingress manifest generation
- Phase 9C: Gateway latency comparison and report enhancements
- Phase 10A: Developer environment doctor
  - Docker, Kubernetes, config, filesystem, Git, and app-path readiness checks
  - safe diagnostics with no automatic installation or app source changes
- Phase 10B: App-aware initialization and safe project integration
  - lightweight top-level project detection
  - relative build-context linking without application source changes
  - generated doctor, validate, and run onboarding steps
- Phase 10C: Interactive guided init
  - opt-in prompt-based configuration with app-aware defaults
  - validated runtime, port, health-path, build, and dependency choices
  - final confirmation before any file is written
- Phase 10D: Onboarding polish and first-run guidance
  - shared app detection and actionable doctor recommendations
  - consistent doctor, validate, explain, and run next steps
  - service-aware interactive image defaults and safer output guidance
- Phase 11: Dashboard structured report UX
  - frontend API and report type parity
  - runtime environment and run timeline review
  - gateway, direct-service, latency, workload, and performance result review
  - compact config summary polish and frontend verification
- Phase 12A: Demo polish and portfolio presentation
  - concise quick demo and architecture guidance
  - sanitized dashboard and structured report screenshots
  - safe demo-data and screenshot capture guidance

Phase 10 is complete. Its onboarding path remains read-only toward application folders: it diagnoses prerequisites, generates only the selected config, and never installs tools or creates application files.

Phase 11 is complete. The dashboard now uses structured JSON reports while preserving Markdown preview and compatibility with older reports that omit newer sections.

Phase 12A is complete. The repository now has a concise local demo flow, refreshed screenshots, and explicit guidance for keeping generated report data out of version control.

### Future

- Optional safe config save-to-file workflow
- GitHub Actions workflow generator

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


