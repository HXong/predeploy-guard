# AGENTS.md

## Project Overview

PreDeploy Guard is a local-first pre-deployment validation and experiment sandbox for backend services.

The CLI validates services by:
1. Loading `predeploy.yaml`
2. Building the target image
3. Creating a temporary Docker Compose sandbox
4. Starting dependencies and target service
5. Running readiness checks
6. Running smoke checks
7. Running Dockerized k6 performance checks
8. Writing Markdown/JSON reports
9. Recording run history

The project also exposes a local API server and React dashboard. It helps developers define services and dependencies, run readiness/smoke/performance checks, and generate local configuration through a guided UI.

The next direction is richer, config-driven deployment experiments across Docker Compose and future Kubernetes-style local runtimes. PreDeploy Guard remains a local sandbox and experiment runner, not a production deployment platform.

## Architecture Rules

### Backend

- Keep `cmd/predeploy` as the CLI adapter only.
- Keep terminal printing in `cmd/predeploy/*_output.go`.
- Keep HTTP handlers in `internal/server`.
- Keep reusable use cases in `internal/app`.
- Do not put CLI printing inside `internal/app`.
- Do not put HTTP response logic inside `internal/app`.
- `internal/app` should return structs/errors, not print or write HTTP responses.
- Do not expose raw `config.Config` through HTTP.
- Expose safe DTOs for API responses, such as config summaries and explanations.
- Keep Docker/k6/filesystem details inside adapter-style packages:
  - `internal/builder`
  - `internal/sandbox`
  - `internal/runtime`
  - `internal/loadtest`
  - `internal/report`
  - `internal/history`

### Runtime and Experiment Guidance

- Keep Docker Compose supported as the default local runtime.
- `internal/runtime` owns runtime-facing lifecycle types and adapter interfaces.
- `internal/runtime/factory` owns runtime adapter selection.
- `internal/runtime/compose` is the only implemented runtime adapter and delegates to the existing Docker Compose sandbox behavior.
- Add future Kubernetes support as a new runtime adapter; do not scatter Kubernetes logic through `internal/runner`.
- Do not hardcode Minikube-specific behavior.
- Future Kubernetes support should use existing kubeconfig contexts and allow Minikube, kind, k3s, and other local development clusters.
- Prefer runtime adapter boundaries over scattering Docker or Kubernetes logic through the runner.
- Preserve existing Docker Compose validation behavior while runtime boundaries evolve.
- Avoid a big-bang runtime refactor.
- Keep cleanup safety as a first-class concern.
- Prefix temporary namespaces and resources clearly, and delete only resources owned by the active run.
- Do not install Docker, Kubernetes, Minikube, or cluster dependencies automatically.
- Do not install or start Minikube automatically.
- Do not make destructive cluster-wide changes.
- Keep production deployment out of scope.
- Keep experiment workloads generic across HTTP load generation, black-box checks, dependency integration checks, file/log replay, background worker jobs, and message producer/consumer experiments.
- Fluent Bit with Kafka may be an example experiment, but do not make the runtime, workload, or configuration architecture specific to Fluent Bit, Kafka, Locust, or any one tool or integration.
- Avoid niche one-off configuration fields unless they generalize to reusable experiment concepts.

### Frontend

- Keep dashboard frontend under `web/dashboard`.
- Use React + TypeScript + Vite.
- Keep `web/dashboard/src/app/App.tsx` thin.
- Keep feature-owned API, types, components, and hooks under `web/dashboard/src/features/`.
- Keep config-builder state, validation, YAML generation, and UI inside `web/dashboard/src/features/config-builder/`.
- Extend the existing config-builder feature for future experiment configuration UI.
- Keep reusable HTTP, component, and utility code under `web/dashboard/src/shared/`.
- Preserve strict TypeScript practices: do not use `React.FC` or `any`, and use explicit prop and event types.
- Do not expose raw backend config through the API.
- Prefer generated YAML preview, copy, and download over backend filesystem writes.
- Any save-to-file flow must be explicit and safe.
- Do not add a UI library unless explicitly requested.
- Keep CSS in `src/styles.css` for now.

## Current Package Responsibilities

### Backend

- `cmd/predeploy`: Cobra CLI commands and terminal output formatting.
- `internal/app`: shared use-case services for CLI and server.
- `internal/config`: load, validate, resolve paths, apply profiles.
- `internal/scaffold`: generate starter `predeploy.yaml` and dependency presets.
- `internal/runner`: orchestrate full validation runs.
- `internal/runtime`: define runtime-neutral lifecycle types and adapter interfaces.
- `internal/runtime/factory`: select the configured runtime adapter.
- `internal/runtime/compose`: adapt the existing Docker Compose sandbox lifecycle, readiness, diagnostics, and cleanup behavior.
- `internal/history`: manage `reports/history.json`.
- `internal/report`: write Markdown and JSON reports.
- `internal/server`: local API server for dashboard.
- `internal/builder`: Docker image build logic.
- `internal/sandbox`: Docker Compose sandbox lifecycle.
- `internal/checker`: readiness and smoke checks.
- `internal/loadtest`: Dockerized k6 execution.

### Frontend

- `web/dashboard/src/app`: thin application shell and top-level view selection.
- `web/dashboard/src/features`: feature-owned dashboard, run, config, and config-builder code.
- `web/dashboard/src/shared`: reusable HTTP, component, and utility code.

## Local API Rules

Current API endpoints:
- `GET /api/health`
- `GET /api/config/summary`
- `GET /api/config/explain`
- `GET /api/history`
- `GET /api/history/{runId}`
- `GET /api/reports/{runId}/json`
- `GET /api/reports/{runId}/markdown`
- `POST /api/runs`
- `GET /api/runs`
- `GET /api/runs/{taskId}`
- `GET /api/runs/{taskId}/logs`

Do not reintroduce:
- `GET /api/config`

Reason: raw config may expose environment values or internal fields.

## Run Task Rules

API-triggered runs are async.

- `POST /api/runs` should return a task ID quickly.
- `GET /api/runs/{taskId}` should return task status.
- `GET /api/runs/{taskId}/logs` should return high-level task logs.
- In-memory run tasks do not replace `reports/history.json`.
- Run history remains the persistent record of completed validation runs.

Current task statuses:
- `queued`
- `running`
- `completed`
- `failed`

## Commands to Run Before Finishing

Backend checks:

```bash
go fmt ./...
go vet ./...
go test ./...
go run ./cmd/predeploy validate examples/predeploy.yaml
go run ./cmd/predeploy explain examples/predeploy.yaml --profile smoke-only
```

Frontend checks:

```bash
cd web/dashboard
npm run build
```

For backend run execution changes:

```bash
go run ./cmd/predeploy run examples/predeploy.yaml --profile smoke-only
```

For history/server changes:

```bash
go run ./cmd/predeploy history examples/predeploy.yaml
go run ./cmd/predeploy serve --config examples/predeploy.yaml
```

Manual API/browser testing may be done by the developer instead of Codex if the environment cannot reliably test local HTTP calls.

## Git / Generated Files

- Do not commit generated reports under `examples/reports/` unless explicitly requested.
- Do not commit temporary test configs like `tmp-predeploy.yaml`.
- Do not commit frontend generated files:
  - `web/dashboard/node_modules/`
  - `web/dashboard/dist/`
  - `*.tsbuildinfo`
- Commit frontend source/config files:
  - `web/dashboard/package.json`
  - `web/dashboard/package-lock.json`
  - `web/dashboard/index.html`
  - `web/dashboard/vite.config.ts`
  - `web/dashboard/tsconfig.json`
  - `web/dashboard/src/**`
- Keep README and ROADMAP updated when phases are completed.
- Keep AGENTS.md updated when architecture rules change.

## Current Roadmap

- Phase 1 completed: Docker Compose validation MVP.
- Phase 2 completed: Dockerized k6, JSON reports, GitHub Actions support.
- Phase 3 completed: config usability, presets, explain, profiles.
- Phase 4 completed: run history, show, compare.
- Phase 5 completed: local API server, safe config endpoints, API-triggered async runs, status polling, basic task logs.
- Phase 6 completed: local React dashboard, run task UI, run details view, report preview, component refactor.
- Phase 7 completed: guided config builder with service, dependency, smoke-check, performance, profile, and YAML export workflows.
- Phase 8A completed: runtime and experiment model alignment.
- Phase 8B current: runtime adapter foundation, with Docker Compose as the only implemented runtime and Kubernetes/local-cluster support reserved for Phase 8C.
