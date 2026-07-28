# Phase 9: Gateway Testing

## Problem

A backend service can behave correctly through its runtime-provided direct URL while failing when traffic passes through an external API gateway or ingress endpoint. Route mapping mistakes, unexpected status handling, and gateway availability can remain invisible to direct smoke and workload checks.

## Goal

Phase 9 adds runtime-neutral validation for gateway-routed traffic. It begins with a user-provided external base URL and selected HTTP routes, then leaves resource generation and latency analysis for later subphases.

## Non-goals

Phase 9 does not:

- install or manage an ingress controller or gateway;
- install NGINX Ingress Controller, Traefik, Kong, Istio, or another gateway;
- add Helm;
- add gateway authentication, headers, or implementation-specific configuration;
- compare gateway and direct-service latency;
- add TLS;
- generate Gateway API resources;
- edit `/etc/hosts` or another host file;
- discover ingress IP addresses automatically.

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

## Phase 9B: Kubernetes Ingress Manifest Generation

Phase 9B optionally generates a `networking.k8s.io/v1` Ingress in the same temporary namespace as the target service:

```yaml
runtime:
  type: kubernetes

gateway:
  enabled: true
  baseURL: http://predeploy.local
  ingress:
    enabled: true
    host: predeploy.local
    className: local-controller
    pathType: Prefix
    annotations:
      example.test/setting: enabled
  routes:
    - name: homepage-via-gateway
      path: /
```

Ingress generation is disabled by default and requires the Kubernetes runtime. The generated resource:

- uses the sanitized target service name and temporary namespace;
- carries the same owned labels as other runtime resources;
- maps each unique configured gateway path to the target ClusterIP Service and service port;
- includes the host, ingress class name, path type, and annotations only as configured;
- is applied in the existing runtime-start manifest operation; and
- is removed with the owned temporary namespace during normal cleanup.

Phase 9B does not create an IngressClass, controller Deployment, LoadBalancer Service, NodePort Service, or any cluster-wide resource. It does not query controller-specific readiness; bounded gateway requests determine whether the generated route is active.

The user remains responsible for installing and configuring a suitable local ingress controller and ensuring `gateway.baseURL` resolves to its endpoint. PreDeploy Guard does not edit host files or discover controller addresses.

### Generated Ingress stabilization

Service readiness can pass before an ingress controller has observed and activated a newly applied route. When `gateway.ingress.enabled` is true, gateway checks therefore retry every second until all configured routes pass or the retry window expires.

The retry window uses `settings.timeoutSeconds` with a maximum of 30 seconds. Only the final route results are included in the report, while the gateway-check phase duration captures the total stabilization wait. Phase 9A checks against externally managed gateways remain single-attempt checks.

## Future Phase 9C: Latency Comparison

Phase 9C may add direct-vs-gateway latency comparisons, thresholds, and report enhancements. These additions should preserve existing report fields and separate gateway overhead analysis from correctness checks.

## Safety Principles

- Treat the gateway as an externally managed URL in Phase 9A.
- Never install, start, reconfigure, or delete a gateway or ingress controller.
- Never make cluster-wide changes.
- Keep gateway checks independent of Docker Compose and Kubernetes adapters.
- Keep generated ingress resources owned and namespace-scoped.
- Preserve configured route order while deduplicating identical paths.
- Send only the configured HTTP method and route; do not expose environment variables or sensitive configuration.
- Use bounded HTTP timeouts and honor run cancellation through context.
- Preserve existing behavior when gateway validation is disabled.
