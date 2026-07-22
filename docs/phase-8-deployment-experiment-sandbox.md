# Phase 8: Deployment Experiment Sandbox

## Problem

Developers often spend significant time manually wiring local deployment environments before they can test realistic behavior. They must connect services to dependencies, prepare representative inputs, start supporting components, run checks, capture results, and clean everything up afterward.

Common examples include:

- service, database, and cache integration testing
- high API traffic experiments
- black-box and input validation testing
- worker and background job testing
- message broker experiments
- log and file replay experiments
- infrastructure component testing, including collectors and sidecars

Fluent Bit with Kafka is one example of this broader pattern, but it is not the architecture's defining use case.

## Goal

PreDeploy Guard should provide a config-driven way to create isolated local deployment experiments, run checks and optional workloads, collect results, and clean up owned resources safely. The existing validation workflow should remain useful while the product grows into a broader local experiment sandbox.

## Non-goals

- deploying to production
- replacing Kubernetes, continuous delivery, or deployment platforms
- installing Minikube, kind, k3s, Docker, Kubernetes, or other cluster dependencies automatically
- becoming specific to Kafka, Fluent Bit, Locust, or any single tool
- performing destructive cluster-wide operations

## Current Runtime

Docker Compose is the current and default supported runtime. It builds the target image, creates a temporary sandbox, starts dependencies and the target service, waits for readiness, runs smoke and Dockerized k6 performance checks, writes Markdown and JSON reports, records history, and cleans up.

Phase 8 must preserve this behavior while runtime boundaries evolve incrementally.

## Future Runtime Direction

Future Kubernetes support should use kubeconfig contexts already configured and controlled by the user. It should work with local targets such as Minikube, kind, k3s, and other local development clusters without assuming that any one distribution is the default.

PreDeploy Guard should not install or start a cluster automatically. A future Kubernetes runtime should create clearly owned, temporary resources, operate within an explicitly selected context, and clean up only resources created for the active run.

## Conceptual Model

### Runtime

The execution backend that creates and manages the sandbox.

Examples:

- `docker-compose`, the current supported runtime
- `kubernetes`, a future runtime using an existing kubeconfig context

### Validation Run

The existing flow that validates a service through dependency and service readiness, smoke checks, and optional performance checks. A validation run uses the current Docker Compose runtime and produces the established reports and history records.

### Deployment Experiment

A broader sandbox run that may include the validation flow plus additional workloads, traffic generation, replay jobs, worker processes, message interactions, or infrastructure components. A deployment experiment remains local-first and does not deploy to production.

### Workload

An optional actor that exercises the sandbox or supplies representative input. Examples include:

- HTTP traffic generator
- log or file replay job
- message producer
- background worker
- custom command job

Workloads should be modeled generically so that core abstractions are not tied to a particular product or protocol.

### Checks

Assertions that evaluate sandbox health or behavior, including readiness, smoke, black-box, input validation, and performance checks. Checks produce pass/fail evidence; workloads generate activity and may have their own result data.

### Reports

The existing Markdown and JSON output, together with future deployment timings, workload results, logs, events, and runtime diagnostics. Report evolution should preserve existing validation report behavior and run history.

## Backend Boundary Direction

Phase 8 should evolve the existing package boundaries gradually instead of replacing the runner in one step:

- `internal/runner` should continue to orchestrate the use case without accumulating runtime-specific Docker or Kubernetes commands.
- `internal/sandbox` is the clearest starting point for a runtime lifecycle boundary: create, start, inspect, and clean up an owned sandbox.
- `internal/builder` should remain responsible for image build details while allowing future runtimes to consume the resulting image through their own mechanisms.
- `internal/checker` and `internal/loadtest` should remain focused on checks and workload execution rather than sandbox provisioning.
- `internal/report` and `internal/history` should consume runtime-neutral results, with runtime diagnostics added incrementally.

The first adapter work should wrap proven Docker Compose behavior rather than rewrite it. Interface extraction should follow concrete needs discovered during Phase 8B and preserve existing CLI, API, report, and cleanup behavior.

## Proposed Phase 8 Subphases

### 8A: Runtime and Experiment Model Alignment

- define runtime, validation run, deployment experiment, workload, checks, and report terminology
- document Docker Compose as the default supported runtime
- identify boundaries that can evolve into runtime adapters
- document future Kubernetes/local-cluster direction without implementing it

### 8B: Runtime Adapter Foundation

- introduce runtime adapter boundaries gradually
- preserve Docker Compose behavior
- avoid a big-bang runner rewrite

### 8C: Kubernetes Local Runtime MVP

- use existing kubeconfig contexts
- support local clusters such as Minikube, kind, and k3s
- create prefixed temporary namespaces and resources
- deploy services and dependencies, wait for readiness, run smoke checks, collect basic diagnostics, and clean up safely

### 8D: Experiment Workloads

- support generic HTTP load, log/file replay, message producer, worker, and custom command workloads
- keep workload concepts independent of any single tool or integration

### 8E: Experiment Reports

- extend Markdown and JSON reports with deployment timings, readiness results, workload results, logs, and runtime diagnostics

## Safety Principles

- remain local-first
- make cleanup explicit and delete only resources owned by the active run
- do not perform production deployment
- do not perform cluster-wide destructive operations
- prefix temporary namespaces and resources clearly
- let the user control the kubeconfig and selected context
- prefer dry-run or preview behavior when possible

## Open Questions

- Should experiment workloads be top-level YAML concepts or part of checks?
- Should Kubernetes manifests be generated for review or applied directly?
- Should the dashboard preview Kubernetes resources before applying them?
- How should logs and replay workloads be modeled generically across files, streams, and protocols?
- Should Phase 8B introduce a runtime adapter interface immediately or evolve the boundary gradually from the existing runner?
