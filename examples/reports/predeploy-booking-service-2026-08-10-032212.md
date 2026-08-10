# PreDeploy Guard Report

## Summary

- **Service:** booking-service
- **Image:** `predeploy-sample-service:latest`
- **Runtime:** docker-compose
- **Base URL:** http://localhost:59444
- **Result:** **PASS**
- **Started At:** 2026-08-10T03:22:12+08:00
- **Finished At:** 2026-08-10T03:22:53+08:00
- **Total Duration:** 40.572s

## Runtime Environment

| Field | Value |
|---|---|
| Runtime | docker-compose |
| Environment | booking-service |
| Base URL | http://localhost:59444 |
| Workload Base URL | http://host.docker.internal:59444 |

## Run Timeline

| Phase | Status | Duration | Error |
|---|---|---:|---|
| image-build | passed | 17.188s | - |
| runtime-prepare | passed | 20ms | - |
| runtime-start | passed | 1.156s | - |
| runtime-readiness | passed | 5.456s | - |
| service-readiness | passed | 40ms | - |
| gateway-checks | skipped | 0ms | - |
| smoke-checks | passed | 14ms | - |
| experiment-workloads | skipped | 0ms | - |
| performance-checks | passed | 16.698s | - |
| diagnostics | skipped | 0ms | - |

## Build Check

| Step | Image | Context | Dockerfile | Result | Error |
|---|---|---|---|---|---|
| docker build | `predeploy-sample-service:latest` | `C:\Users\ongho\predeploy-guard\examples\sample-service` | `Dockerfile` | PASS | - |

## Readiness Checks

| Check | Target | Result | Error |
|---|---|---|---|
| dependency readiness | postgres | PASS | - |
| dependency readiness | redis | PASS | - |
| service readiness | http://localhost:59444/health | PASS | - |

## Gateway Checks

No gateway checks were executed.

## Smoke Checks

| Check | Method | URL | Expected | Actual | Duration | Result | Error |
|---|---|---|---:|---:|---:|---|---|
| health check | GET | http://localhost:59444/health | 200 | 200 | 9ms | PASS | - |
| list bookings | GET | http://localhost:59444/api/bookings | 200 | 200 | 5ms | PASS | - |

## Experiment Workloads

No experiment workloads were configured.

## Performance Check

### Summary

- VUs: `10`
- Duration: `15s`
- Result: **PASS**

### Latency Metrics

| Metric | Value | Threshold | Result |
|---|---:|---:|---|
| avg latency | 7.87ms | - | - |
| min latency | 1.13ms | - | - |
| median latency | 6.99ms | - | - |
| max latency | 31.58ms | - | - |
| p90 latency | 14.52ms | - | - |
| p95 latency | 16.87ms | 300.00ms | PASS |

### Reliability Metrics

| Metric | Value | Threshold | Result |
|---|---:|---:|---|
| error rate | 0.0000 | 0.0100 | PASS |
| request count | 300 | - | - |
| iterations | 150 | - | - |
| checks total | 0 | - | - |
| check pass rate | 1.0000 | 1.0000 | PASS |

## Check Summary

- Readiness checks: 3 total, 3 passed, 0 failed
- Gateway checks: 0 total, 0 passed, 0 failed
- Smoke checks: 2 total, 2 passed, 0 failed
- Experiment workloads: 0 configured

## Configuration

- Cleanup enabled: `true`
- Timeout seconds: `60`
- Health path: `/health`

