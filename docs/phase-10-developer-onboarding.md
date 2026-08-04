# Phase 10: Developer Onboarding & App Integration

## Problem

Connecting an existing backend service to a local deployment experiment currently assumes that the developer already knows which Docker, Kubernetes, configuration, and application prerequisites are available. Generating integration files before checking those prerequisites would make onboarding harder to trust and could risk changing a user project unnecessarily.

## Goal

Phase 10 makes onboarding safe and incremental. It starts with diagnostics that explain local readiness, then adds explicit app-aware configuration generation backed by lightweight project detection. PreDeploy Guard remains a local validation sandbox and does not install platform tools or alter application source code.

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

## Phase 10B: App-Aware Init and Safe Project Integration

Phase 10B adds an explicit way to create a reviewable PreDeploy Guard config for an existing application directory:

```bash
predeploy init --app ./my-app
predeploy init --app ./my-app --port 8080 --health-path /health
predeploy init --app ./my-app --runtime kubernetes
predeploy init --app ./my-app --service-name orders-api
predeploy init --app ./my-app --image orders-api:local
predeploy init --app ./my-app --no-build
```

### Purpose

App-aware init reduces the manual edits needed to connect a local project while keeping every generated decision visible and overridable. It derives a sanitized service name from the app directory and defaults to:

- runtime `docker-compose`;
- image `predeploy-<service-name>:local`;
- port `8080`; and
- health path `/health`.

The runtime, service name, image, port, and health path can all be selected explicitly. The normal `predeploy init` output remains unchanged when `--app` is not supplied.

### Generated config behavior

Detection checks only the app directory's top-level `Dockerfile`, `package.json`, `go.mod`, `requirements.txt`, `pyproject.toml`, `pom.xml`, and `build.gradle` filenames. It does not read their contents.

When a Dockerfile is present and `--no-build` is not selected, the generated service includes a build context relative to the output config directory where possible:

```yaml
runtime:
  type: docker-compose

service:
  name: "my-app"
  image: "predeploy-my-app:local"
  build:
    context: "./my-app"
    dockerfile: Dockerfile
  port: 8080
  healthPath: "/health"
```

Without a Dockerfile, the config keeps the image reference and omits `service.build`. Init prints a warning explaining that the image must already exist or a Dockerfile can be added later. `--no-build` deliberately produces the same image-only configuration even when a Dockerfile exists.

Generated app-aware configs include a health smoke check using the selected health path, keep gateway and performance checks disabled by default, and retain the existing smoke-only, light-load, and stress-test profiles. The completion output recommends:

```bash
predeploy doctor --config predeploy.yaml --app ./my-app
predeploy validate predeploy.yaml
predeploy run predeploy.yaml
```

### Safety boundaries

- The app directory must already exist and must be a directory.
- Init writes only the selected output config and rejects an output path inside the app directory.
- Existing output is preserved unless `--force` is supplied.
- No application file is created, changed, or deleted.
- No Dockerfile is generated.
- `.env` files, dependency manifests, and source files are never opened or parsed.
- No tool is installed or started, and no host file is edited.
- An invalid runtime or missing app path fails before config output is written.
- Relative build contexts are used when the app and config locations permit it.

## Phase 10C: Interactive Guided Init

Phase 10C adds an opt-in prompt-driven setup flow for developers who prefer to confirm configuration choices interactively:

```bash
predeploy init --interactive
predeploy init --interactive --app ./my-app
predeploy init --interactive --output predeploy.yaml
predeploy init --interactive --app ./my-app --runtime kubernetes --port 8080
```

Without `--interactive`, both generic and app-aware init retain their existing non-interactive behavior and never display prompts.

### Purpose

Guided init makes the safe Phase 10B generator easier to discover without adding hidden inference. Existing flags become prompt defaults, so a developer can preselect some answers and confirm or adjust the rest. A blank app directory creates a generic starter; an app path activates the existing lightweight top-level detection.

### Prompt flow

The wizard asks for:

1. application directory;
2. service name;
3. Docker Compose or Kubernetes runtime;
4. service image;
5. container port;
6. health path;
7. Dockerfile build-context use when a top-level Dockerfile exists;
8. optional PostgreSQL and Redis dependency presets;
9. output config path; and
10. final confirmation.

Detected app indicators and the suggested service name are shown before service prompts. Runtime, port, health-path, yes/no, and dependency choices are validated and re-prompted rather than silently corrected. The final preview includes the output, runtime, service, image, build context, port, health path, and dependencies.

The wizard itself only reads answers and returns scaffold options. The CLI calls the existing scaffold writer only after the developer confirms. Answering no at the final prompt prints `Init cancelled. No files were written.` and returns successfully.

### Safety boundaries

- Prompts appear only when `--interactive` is explicitly supplied.
- The wizard uses injected input and output streams and performs no direct terminal or file writes.
- App detection must succeed before the configuration prompts continue.
- No config is written before final confirmation.
- Cancellation and detection failures write no files.
- Existing output remains protected unless `--force` was supplied before the wizard started.
- The generated config must remain outside the detected application directory.
- The application directory, source files, Dockerfiles, and `.env` files are never changed or parsed.
- No external tool is installed, started, or reconfigured.

## Phase 10D: Onboarding Polish and First-Run Guidance

Phase 10D connects doctor, guided init, validation, and execution into one consistent first-run path. It does not change runtime orchestration or configuration semantics.

### Before

The onboarding commands existed independently, so developers had to infer whether to initialize first, rerun doctor with the generated config, explain the plan, or move directly to a run. Doctor also maintained a duplicate list of application indicator filenames.

### After

The recommended flow for an existing application is explicit:

```bash
predeploy doctor --app ./my-app
predeploy init --interactive --app ./my-app
predeploy doctor --config predeploy.yaml --app ./my-app
predeploy validate predeploy.yaml
predeploy explain predeploy.yaml
predeploy run predeploy.yaml
```

Doctor now reuses `internal/appdetect` and adds recommendations based on observed state:

- missing config: run generic or app-aware guided init;
- valid config: validate and run the selected config;
- invalid config: fix validation errors before running;
- unavailable configured Kubernetes runtime: check kubeconfig and context;
- generated ingress: provide an ingress controller externally and ensure `gateway.baseURL` resolves to it; and
- enabled performance checks without Docker: make the Docker CLI and daemon available for Dockerized k6.

Recommendations are printed commands, not automatic actions. They contain only user-provided paths and fixed command names, never environment values or config secrets.

Init output now presents doctor, validate, explain, and run in the same order for app-aware and interactive generation. Generic init adds config review first. Profile, history, show, and compare commands remain available after the first-run steps.

Interactive image defaults follow the final service name unless `--image` was supplied. For example, choosing `orders-api` suggests `orders-api:predeploy` for a generic starter or `predeploy-orders-api:local` for an app-aware starter.

When an output path points inside the app directory, init still refuses to write and now shows a safe command targeting a config adjacent to the app.

### Safety boundaries

- Doctor recommendations never execute commands or install tools.
- Recommended commands never include environment values or application file contents.
- Shared app detection inspects only known top-level filenames.
- Init continues to write only the selected config outside the app directory.
- No application source, Dockerfile, dependency manifest, or `.env` file is read or modified.
- Existing command flags and non-interactive behavior remain compatible.

## Future Improvements

Future onboarding work may add terminal-aware selection controls, richer previews, or carefully scoped detection-informed suggestions. Any enhancement must preserve stream-injected tests, explicit overrides, non-interactive automation, final confirmation, and the no-application-modification boundary.

## Non-goals

Phase 10 does not:

- install Docker, Docker Compose, kubectl, Kubernetes, Minikube, k6, or another tool;
- install, select, or manage an ingress controller;
- start Docker, Minikube, or a Kubernetes cluster;
- run a validation experiment or k6 workload;
- edit host files, kubeconfig, Docker configuration, or user application files;
- generate a Dockerfile, Compose file, Kubernetes manifest, or any artifact inside the application directory;
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
