export type EnvVarDraft = {
  id: string;
  key: string;
  value: string;
};

export type RuntimeType = "docker-compose";

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
  settings: {
    cleanup: boolean;
    timeoutSeconds: number;
  };
};

export type ValidationError = {
  field: string;
  message: string;
};
