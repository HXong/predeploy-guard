import type { ConfigBuilderState, DependencyDraft, EnvVarDraft, SmokeCheckDraft } from "./types";

let dependencySequence = 0;
let envSequence = 0;
let smokeCheckSequence = 0;

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
    smoke: [],
  },
  settings: {
    cleanup: true,
    timeoutSeconds: 60,
  },
};

export function createSmokeCheck(): SmokeCheckDraft {
  smokeCheckSequence += 1;
  return {
    id: `smoke-${Date.now()}-${smokeCheckSequence}`,
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
