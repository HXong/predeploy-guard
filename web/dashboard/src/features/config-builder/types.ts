export type EnvVarDraft = {
  id: string;
  key: string;
  value: string;
};

export type RuntimeType = "docker-compose";

export type ReadinessMode = "none" | "shell" | "command";

export type DependencyReadinessDraft = {
  mode: ReadinessMode;
  shell: string;
  command: string;
  intervalSeconds: number;
  timeoutSeconds: number;
};

export type DependencyDraft = {
  id: string;
  name: string;
  image: string;
  port?: number;
  env: EnvVarDraft[];
  readiness: DependencyReadinessDraft;
};

export type ConfigBuilderState = {
  runtime: {
    type: RuntimeType;
  };
  service: {
    name: string;
    image: string;
    build: {
      context: string;
      dockerfile: string;
    };
    port: number;
    healthPath: string;
    env: EnvVarDraft[];
  };
  dependencies: DependencyDraft[];
  settings: {
    cleanup: boolean;
    timeoutSeconds: number;
  };
};

export type ValidationError = {
  field: string;
  message: string;
};
