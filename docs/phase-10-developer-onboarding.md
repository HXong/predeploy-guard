# Phase 10: Developer Onboarding & App Integration

## Problem

Connecting an existing backend service to a local deployment experiment currently assumes that the developer already knows which Docker, Kubernetes, configuration, and application prerequisites are available. Generating integration files before checking those prerequisites would make onboarding harder to trust and could risk changing a user project unnecessarily.

## Goal

Phase 10 makes onboarding safe and incremental. It starts with diagnostics that explain local readiness, then can evolve toward explicit configuration generation and lightweight project detection. PreDeploy Guard remains a local validation sandbox and does not install platform tools or alter application source code.

## Phase 10A: Developer Environment Doctor

`predeploy doctor` performs a general readiness scan:

```bash
predeploy doctor
predeploy doctor --config predeploy.yaml
predeploy doctor --app ./my-app
predeploy doctor --config predeploy.yaml --app ./my-app
```

The report groups checks into local environment, Docker, Kubernetes, and optional application sections. Each check is marked `PASS`, `WARN`, or `FAIL`.

The local environment checks cover:

- current-directory writability through a temporary file that is immediately removed;
- `.git` detection in the current directory or its parents; and
- default or explicitly selected PreDeploy Guard configuration validation.

The Docker checks cover:

- Docker CLI discovery;
- Docker daemon reachability through `docker info`; and
- Docker Compose availability through `docker compose version`.

The Kubernetes checks cover:

- kubectl discovery;
- the current kubeconfig context;
- cluster reachability through `kubectl cluster-info`; and
- optional Minikube discovery.

Minikube is never required by itself. Any existing, accessible kubeconfig context can support the Kubernetes runtime.

When `--app` is supplied, doctor verifies that the path is a directory and looks only for these top-level indicators:

- `Dockerfile`
- `package.json`
- `go.mod`
- `requirements.txt`
- `pyproject.toml`
- `pom.xml`
- `build.gradle`

It does not parse those files. A missing Dockerfile is a warning because an existing image can still be configured, while future build-context integration will need a Dockerfile.

### Config-aware requirements

When a config loads successfully, doctor shows its configured runtime and makes the relevant capabilities required:

- Docker CLI, daemon, and Compose for `docker-compose`;
- kubectl, context, and cluster reachability for `kubernetes`;
- kubectl and cluster access when generated ingress is enabled; and
- Docker CLI and daemon when Dockerized performance checks are enabled.

Generated ingress diagnostics explicitly remind developers that PreDeploy Guard does not install or manage an ingress controller. Performance diagnostics do not run k6.

Without a valid config, missing optional tools are warnings so a Docker-only or Kubernetes-only developer can still use the report. Warnings return exit code 0. A failed filesystem, app-path, config, or required capability check returns exit code 1.

## Future Phase 10B: `init --app`

A future `predeploy init --app <folder>` workflow may generate a reviewable integration config for an existing project. It must be explicit about output paths and overwrite behavior. It must not modify application source, generate app-specific Dockerfiles without a separate design, or silently install dependencies.

## Future Phase 10C: Project Detection

Future project detection may infer a small set of useful defaults from common top-level files. Detection should remain lightweight, transparent, and easy to override. Deep framework detection, package-manager inspection, and environment discovery are intentionally deferred.

## Non-goals

Phase 10A does not:

- install Docker, Docker Compose, kubectl, Kubernetes, Minikube, k6, or another tool;
- install, select, or manage an ingress controller;
- start Docker, Minikube, or a Kubernetes cluster;
- run a validation experiment or k6 workload;
- edit host files, kubeconfig, Docker configuration, or user application files;
- generate `predeploy.yaml`, a Dockerfile, Compose files, or Kubernetes manifests;
- deeply detect frameworks or parse dependency manifests; or
- read or print environment values or secret-bearing files.

## Safety Principles

- Diagnose before generating or modifying integration artifacts.
- Treat missing tools as optional unless a loaded configuration requires them.
- Use existing kubeconfig contexts without assuming Minikube or a specific Kubernetes distribution.
- Keep subprocesses read-only and bounded by the command context.
- Inspect only paths and known top-level filenames for application checks.
- Never expose configuration environment values or file contents.
- Remove the temporary writability probe immediately and report cleanup failures.
- Never install software or change host and application configuration.
