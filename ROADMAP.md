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
Goal: allow developers to define and run local deployment experiments from `predeploy.yaml`, using Docker Compose by default and existing Kubernetes/local-cluster contexts through the runtime adapter architecture.

PreDeploy Guard remains a local pre-deployment sandbox and experiment runner. It is not a production deployment platform.

#### Phase 8A: Runtime and Experiment Model Alignment (Completed)
- Clarify the runtime abstraction direction
- Keep Docker Compose as the default runtime
- Define how deployment experiments differ from normal validation runs
- Document intended support for future Kubernetes and local-cluster execution
- Do not implement Kubernetes in this subphase

#### Phase 8B: Runtime Adapter Foundation (Completed)
- Introduce runtime-neutral lifecycle types and an adapter interface
- Add runtime adapter selection with Docker Compose as the first and default implementation
- Adapt the existing Docker Compose sandbox lifecycle without duplicating it
- Keep high-level validation orchestration, checks, reports, and history in their existing packages
- Preserve existing Docker Compose behavior
- Avoid a big-bang refactor

#### Phase 8C: Kubernetes Local Runtime MVP (Completed)
- Use existing kubeconfig contexts
- Support local clusters such as Minikube, kind, or k3s
- Use `kubectl` without adding Kubernetes client dependencies
- Create clearly owned temporary namespaces
- Generate and apply service/dependency manifests
- Wait for deployment readiness and start a local port-forward
- Run the existing smoke and performance checks through the forwarded URL
- Collect basic logs and runtime diagnostics
- Clean up only the owned namespace and temporary files
- Require users to make configured images accessible to the selected cluster

#### Phase 8D: Experiment Workloads (Current)
- Implement runtime-neutral HTTP workloads first
- Support configurable method, path, duration, request rate, expected status, and `fail` or `warn` failure policy
- Run workloads through the runtime-provided base URL after smoke checks and before k6
- Include structured workload results in Markdown and JSON reports
- Keep log/file replay, message producers/consumers, worker jobs, custom commands, and Kubernetes Jobs as future extensions
- Do not specialize around Fluent Bit, Kafka, Locust, or any single tool

#### Phase 8E: Experiment Reports
- Extend reports with deployment timings, readiness results, workload results, logs, and runtime diagnostics

## Future Enhancements

- Optional safe save-to-file flow through the local API
- GitHub Actions workflow generator for selected configurations and profiles

## Future Direction

### Phase 9: API Gateway / Ingress Testing
- Gateway config support
- Direct service vs gateway latency comparison
