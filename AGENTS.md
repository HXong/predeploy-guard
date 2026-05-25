# AGENTS.md

## Project Overview

PreDeploy Guard is a local-first pre-deployment validation tool for backend services.

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

The project also exposes a local API server for future GUI/dashboard support.

Future direction: GDP-inspired GUI/platform mode built on top of the CLI engine.

## Architecture Rules

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

## Current Package Responsibilities

- `cmd/predeploy`: Cobra CLI commands and terminal output formatting.
- `internal/app`: shared use-case services for CLI and server.
- `internal/config`: load, validate, resolve paths, apply profiles.
- `internal/scaffold`: generate starter `predeploy.yaml` and dependency presets.
- `internal/runner`: orchestrate full validation runs.
- `internal/history`: manage `reports/history.json`.
- `internal/report`: write Markdown and JSON reports.
- `internal/server`: local API server for future GUI.
- `internal/builder`: Docker image build logic.
- `internal/sandbox`: Docker Compose sandbox lifecycle.
- `internal/checker`: readiness and smoke checks.
- `internal/loadtest`: Dockerized k6 execution.

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

```bash
go fmt ./...
go vet ./...
go test ./...
go run ./cmd/predeploy validate examples/predeploy.yaml
go run ./cmd/predeploy explain examples/predeploy.yaml --profile smoke-only
```

### For changes involving run execution

```bash
go run ./cmd/predeploy run examples/predeploy.yaml --profile smoke-only
```

### For changes involving history/server

```bash
go run ./cmd/predeploy history examples/predeploy.yaml
go run ./cmd/predeploy serve --config examples/predeploy.yaml
```

Manual API testing may be done by the developer instead of Codex if the environment cannot reliably test local HTTP calls.

## Git / Generated Files

- Do not commit generated reports under `examples/reports/` unless explicitly requested.
- Do not commit temporary test configs like `tmp-predeploy.yaml`.
- Keep README and ROADMAP updated when phases are completed.
- Keep AGENTS.md updated when architecture rules change.

## Current Roadmap

- Phase 1 completed: Docker Compose validation MVP.
- Phase 2 completed: Dockerized k6, JSON reports, GitHub Actions support.
- Phase 3 completed: config usability, presets, explain, profiles.
- Phase 4 completed: run history, show, compare.
- Phase 5 completed: local API server, safe config endpoints, API-triggered async runs, status polling, basic task logs.
- Phase 6 current/next: local dashboard UI.
