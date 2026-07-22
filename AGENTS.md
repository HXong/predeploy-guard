# AGENTS.md

## Project Overview

PreDeploy Guard is a local-first pre-deployment validation platform for backend services.

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

The project also exposes a local API server and React dashboard for future GUI/platform workflows.

Future direction: GDP-inspired local platform mode built on top of the CLI engine.

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
  - `internal/loadtest`
  - `internal/report`
  - `internal/history`

### Frontend

- Keep dashboard frontend under `web/dashboard`.
- Use React + TypeScript + Vite.
- Keep `web/dashboard/src/app/App.tsx` thin.
- Keep feature-owned API, types, components, and hooks under `web/dashboard/src/features/`.
- Keep config-builder state, validation, YAML generation, and UI inside `web/dashboard/src/features/config-builder/`.
- Keep reusable HTTP, component, and utility code under `web/dashboard/src/shared/`.
- Preserve strict TypeScript practices: do not use `React.FC` or `any`, and use explicit prop and event types.
- Do not expose raw backend config through the API.
- Prefer generated YAML preview, copy, and download over backend filesystem writes unless explicitly implementing a safe save flow.
- Do not add a UI library unless explicitly requested.
- Keep CSS in `src/styles.css` for now.

## Current Package Responsibilities

### Backend

- `cmd/predeploy`: Cobra CLI commands and terminal output formatting.
- `internal/app`: shared use-case services for CLI and server.
- `internal/config`: load, validate, resolve paths, apply profiles.
- `internal/scaffold`: generate starter `predeploy.yaml` and dependency presets.
- `internal/runner`: orchestrate full validation runs.
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
- Phase 8 current/next: GitHub Actions workflow generator.
