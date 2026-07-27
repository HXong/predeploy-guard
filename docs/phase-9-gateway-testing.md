# Phase 9: Gateway Testing

## Problem

A backend service can behave correctly through its runtime-provided direct URL while failing when traffic passes through an external API gateway or ingress endpoint. Route mapping mistakes, unexpected status handling, and gateway availability can remain invisible to direct smoke and workload checks.

## Goal

Phase 9 adds runtime-neutral validation for gateway-routed traffic. It begins with a user-provided external base URL and selected HTTP routes, then leaves resource generation and latency analysis for later subphases.

## Non-goals

Phase 9A does not:

- install or manage an ingress controller or gateway;
- install NGINX Ingress Controller, Traefik, Kong, Istio, or another gateway;
- generate Kubernetes Ingress or gateway resources;
- add Helm;
- change Docker Compose or Kubernetes runtime behavior;
- add gateway authentication, headers, or implementation-specific configuration;
- compare gateway and direct-service latency.

## Phase 9A: External Gateway Checks

Users can enable gateway validation in `predeploy.yaml`:

```yaml
gateway:
  enabled: true
  baseURL: http://localhost:8088
  routes:
    - name: homepage-via-gateway
      method: GET
      path: /
      expectedStatus: 200
      compareDirect: true
```

Gateway checks run after service readiness and before smoke checks. Each configured route is requested through `gateway.baseURL`. When `compareDirect` is enabled, the same method and path are also requested through the direct runtime service URL in `env.BaseURL`.

A compared route passes only when:

1. the gateway request succeeds with the expected status;
2. the direct request succeeds with the expected status; and
3. the gateway and direct statuses match.

An omitted `method` defaults to `GET`, an omitted `expectedStatus` defaults to `200`, and an omitted `compareDirect` defaults to `true`. An explicit `compareDirect: false` validates only the gateway response.

If any gateway route fails, the run fails and smoke, workload, and performance checks are skipped. When gateway checks are disabled, existing run behavior is unchanged.

Markdown and JSON reports include structured gateway results. The Markdown gateway table appears after readiness checks and before smoke checks.

## Future Phase 9B: Ingress/Gateway Resource Generation

Phase 9B may explore generating clearly owned Kubernetes Ingress manifests or adding safe gateway runtime integration. Any design must remain portable across local development clusters and avoid binding the configuration model to one controller or gateway product.

## Future Phase 9C: Latency Comparison

Phase 9C may add direct-vs-gateway latency comparisons, thresholds, and report enhancements. These additions should preserve existing report fields and separate gateway overhead analysis from correctness checks.

## Safety Principles

- Treat the gateway as an externally managed URL in Phase 9A.
- Never install, start, reconfigure, or delete a gateway or ingress controller.
- Never make cluster-wide changes.
- Keep gateway checks independent of Docker Compose and Kubernetes adapters.
- Send only the configured HTTP method and route; do not expose environment variables or sensitive configuration.
- Use bounded HTTP timeouts and honor run cancellation through context.
- Preserve existing behavior when gateway validation is disabled.
