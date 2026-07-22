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
- `predeploy serve`
- Safe config summary/explain endpoints
- History/report API endpoints
- Async API-triggered validation runs
- In-memory run task status tracking
- Basic run task logs
- Server route tests

### Phase 6: Local Dashboard UI
Goal: build a simple local dashboard on top of the local API.

Completed:
- React + TypeScript + Vite dashboard under `web/dashboard`
- Vite dev proxy to local API server
- API health card
- Config summary card
- Execution plan summary
- Trigger run panel with profile selector
- Async run task table
- Task status polling
- Task logs panel
- Run history table
- Run details panel
- Markdown/JSON report links
- Raw Markdown report preview
- Dashboard component refactor
- Frontend API client and typed DTOs

### Phase 7: Guided Config Builder
Goal: allow users to generate and edit `predeploy.yaml` through a guided UI instead of writing YAML manually.

Completed:
- Service configuration form
- Dependency wizard with PostgreSQL/Redis presets and readiness configuration
- HTTP smoke check builder
- Base k6 performance builder with thresholds and endpoint fallback
- Validation profile builder with performance presets and overrides
- YAML preview panel
- Copy/download generated `predeploy.yaml`

## Current Phase

### Phase 8: Deployment Experiment Sandbox
Goal: allow developers to define and run local deployment experiments from `predeploy.yaml`, using Docker Compose first and preparing the architecture for future Kubernetes/local-cluster runtimes.

PreDeploy Guard remains a local pre-deployment sandbox and experiment runner. It is not a production deployment platform.

#### Phase 8A: Runtime and Experiment Model Alignment
- Define experiment and runtime terminology
- Document the current Docker Compose runtime as the default supported runtime
- Document the future Kubernetes/local-cluster direction
- Identify backend boundaries that should evolve into runtime adapters
- No Kubernetes implementation yet

#### Phase 8B: Runtime Adapter Foundation
- Introduce runtime adapter boundaries gradually
- Preserve Docker Compose behavior
- Avoid a big-bang runner rewrite

#### Phase 8C: Kubernetes Local Runtime MVP
- Use existing kubeconfig contexts
- Support local clusters such as Minikube, kind, or k3s
- Create temporary namespaces
- Deploy service/dependencies
- Wait for readiness
- Run smoke checks
- Collect basic logs/events
- Clean up safely

#### Phase 8D: Experiment Workloads
- Support generic workloads such as HTTP load, log/file replay, message producers, worker jobs, or custom commands
- Do not specialize around Fluent Bit/Kafka only

#### Phase 8E: Experiment Reports
- Extend reports with deployment timings, readiness results, workload results, logs, and runtime diagnostics

## Future Enhancements

- Optional safe save-to-file flow through the local API
- GitHub Actions workflow generator for selected configurations and profiles

## Future Direction

### Phase 9: API Gateway / Ingress Testing
- Gateway config support
- Direct service vs gateway latency comparison
