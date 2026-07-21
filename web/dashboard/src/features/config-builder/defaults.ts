import type {
  ConfigBuilderState,
  DependencyDraft,
  EnvVarDraft,
  PerformanceEndpointDraft,
  SmokeCheckDraft,
} from "./types";

let dependencySequence = 0;
let envSequence = 0;
let smokeCheckSequence = 0;
let performanceEndpointSequence = 0;

export const defaultConfigBuilderState: ConfigBuilderState = {
  runtime: {
    type: "docker-compose",
  },
  service: {
    name: "my-service",
    image: "my-service:predeploy",
    build: {
      context: ".",
      dockerfile: "Dockerfile",
    },
    port: 8080,
    healthPath: "/health",
    env: [
      {
        id: "env-port",
        key: "PORT",
        value: "8080",
      },
    ],
  },
  dependencies: [],
  checks: {
    smoke: [createHealthCheck()],
  },
  performance: {
    enabled: true,
    vus: 10,
    duration: "15s",
    thresholds: {
      maxP95LatencyMs: 300,
      maxErrorRate: 0.01,
    },
    endpoints: [createHealthPerformanceEndpoint()],
  },
  settings: {
    cleanup: true,
    timeoutSeconds: 60,
  },
};

export function createBlankPerformanceEndpoint(): PerformanceEndpointDraft {
  return {
    id: nextPerformanceEndpointId(),
    name: "",
    method: "GET",
    path: "",
  };
}

export function createHealthPerformanceEndpoint(): PerformanceEndpointDraft {
  return {
    id: nextPerformanceEndpointId(),
    name: "health check load",
    method: "GET",
    path: "/health",
  };
}

export function createBlankSmokeCheck(): SmokeCheckDraft {
  return {
    id: nextSmokeCheckId(),
    name: "",
    method: "GET",
    path: "",
    expectedStatus: 200,
  };
}

export function createHealthCheck(): SmokeCheckDraft {
  return {
    id: nextSmokeCheckId(),
    name: "health check",
    method: "GET",
    path: "/health",
    expectedStatus: 200,
  };
}

export function createCustomDependency(): DependencyDraft {
  return {
    id: nextDependencyId("dependency"),
    name: "",
    image: "",
    port: undefined,
    env: [],
    readiness: {
      mode: "none",
      shell: "",
      command: "",
      intervalSeconds: 2,
      timeoutSeconds: 30,
    },
  };
}

export function createPostgresDependency(): DependencyDraft {
  return {
    id: nextDependencyId("postgres"),
    name: "postgres",
    image: "postgres:16",
    port: 5432,
    env: [
      createEnvVar("POSTGRES_USER", "test"),
      createEnvVar("POSTGRES_PASSWORD", "test"),
      createEnvVar("POSTGRES_DB", "testdb"),
    ],
    readiness: {
      mode: "command",
      shell: "",
      command: "pg_isready -U test -d testdb",
      intervalSeconds: 2,
      timeoutSeconds: 30,
    },
  };
}

export function createRedisDependency(): DependencyDraft {
  return {
    id: nextDependencyId("redis"),
    name: "redis",
    image: "redis:7",
    port: 6379,
    env: [],
    readiness: {
      mode: "shell",
      shell: "redis-cli ping | grep PONG",
      command: "",
      intervalSeconds: 2,
      timeoutSeconds: 30,
    },
  };
}

export function createEnvVar(key = "", value = ""): EnvVarDraft {
  envSequence += 1;
  return {
    id: `env-${Date.now()}-${envSequence}`,
    key,
    value,
  };
}

function nextDependencyId(prefix: string): string {
  dependencySequence += 1;
  return `${prefix}-${Date.now()}-${dependencySequence}`;
}

function nextSmokeCheckId(): string {
  smokeCheckSequence += 1;
  return `smoke-${Date.now()}-${smokeCheckSequence}`;
}

function nextPerformanceEndpointId(): string {
  performanceEndpointSequence += 1;
  return `performance-endpoint-${Date.now()}-${performanceEndpointSequence}`;
}
