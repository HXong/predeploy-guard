# Phase 11: Dashboard Structured Report UX (Completed)

Phase 11 brings the React dashboard into parity with the runtime-neutral configuration and JSON report data added in earlier phases. It improves review before and after a run without changing validation, runtime, gateway, workload, or performance execution behavior.

## Goal

The dashboard should answer two practical questions:

- Before a run: what service, runtime, gateway, workload, and performance configuration is selected?
- After a run: what environment was created, which phases ran, and what did the configured checks report?

## Dashboard Capabilities

The config summary presents compact, safe DTO data for:

- service and runtime selection;
- configured profiles, dependencies, and smoke checks;
- gateway routes, ingress state, and gateway latency policy;
- workload count; and
- performance state and endpoint count.

Selecting a history entry loads its structured JSON report and can display:

- runtime environment details;
- the run phase timeline;
- gateway and optional direct-service results;
- gateway latency, overhead, warnings, and errors;
- workload configuration, counts, status distributions, policy, and outcome; and
- performance latency, thresholds, reliability metrics, and outcome.

Markdown and JSON report links remain available, and the existing raw Markdown preview remains unchanged.

## Compatibility

Newer report sections are optional in the frontend model. Reports created before Phase 8E or later report additions remain reviewable: missing sections show a neutral empty state instead of crashing the dashboard. Missing individual values render as `-`, and disabled or unexecuted checks are not presented as successful checks.

The frontend preserves the backend's existing JSON field names, including the mixed casing of legacy report fields and untagged workload result fields.

## Safety Boundaries

The dashboard remains local, read-only toward reports and configuration summaries, and runtime-neutral. It does not:

- expose configured environment variable values, secrets, cluster-wide data, or application file contents;
- dump raw performance output or runtime diagnostics into the default structured view;
- execute doctor or onboarding commands;
- install or configure Docker, Kubernetes tools, gateways, or ingress controllers;
- modify application folders; or
- alter runtime or experiment behavior.

## Intentionally Not Included

Phase 11 does not add report mutation, production deployment controls, runtime diagnostics display, doctor execution, onboarding forms, or tool installation. Config Builder behavior and the backend report format remain unchanged.
