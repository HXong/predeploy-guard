# PreDeploy Guard Roadmap

## Completed

### Phase 1: Docker Compose Validation MVP
- Docker Compose sandbox
- Target image build
- Dependency services
- Dependency readiness checks
- Service readiness checks
- Smoke checks
- Markdown reports

### Phase 2: Performance and Automation
- Dockerized k6 performance checks
- p95 latency and error-rate thresholds
- Rich performance metrics
- JSON reports
- Config-relative path resolution
- GitHub Actions workflow support

### Phase 3: Config Usability
- `predeploy init`
- Dependency presets
- `predeploy validate`
- `predeploy explain`
- Validation profiles

### Phase 4: Run History and Comparison
- `reports/history.json`
- `predeploy history`
- `predeploy show`
- `predeploy compare`

### Phase 5: Local API and GUI Groundwork
Goal: introduce a local HTTP API so the future GUI can reuse the existing CLI engine.

Completed:
- `predeploy serve`
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
- Async API-triggered validation runs
- In-memory run task status tracking
- Basic run task logs
- Server route tests

## Current Phase

### Phase 6: Local Dashboard UI
Goal: build a simple local dashboard on top of the existing API.

Planned:
- React + TypeScript frontend
- API health card
- Config summary card
- Run history table
- Run task table
- Run button with profile selector
- Poll `GET /api/runs/{taskId}` for task status
- Show task logs from `GET /api/runs/{taskId}/logs`
- Link to Markdown/JSON report endpoints

## Future Direction

### Phase 7: Guided Config Builder
- Service config wizard
- Dependency wizard
- Smoke check builder
- k6 profile builder
- GitHub Actions workflow generator

### Phase 8: Kubernetes Runtime
- Existing cluster support
- Temporary namespace creation
- Manifest deployment
- Helm chart support

### Phase 9: API Gateway / Ingress Testing
- Gateway config support
- Direct service vs gateway latency comparison
