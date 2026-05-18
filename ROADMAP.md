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

## Current Phase

### Phase 5: Local API and GUI Groundwork
Goal: introduce a local HTTP API so the future GUI can reuse the existing CLI engine.

Planned:
- `predeploy serve`
- `GET /api/health`
- `GET /api/history`
- `GET /api/history/{runId}`
- `GET /api/reports/{runId}/json`
- `GET /api/reports/{runId}/markdown`
- Config summary endpoints
- Run trigger endpoint later

## Future Direction

### Phase 6: Local GUI
- React + TypeScript frontend
- Go local API backend
- Report dashboard
- Run history view
- Profile comparison view

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