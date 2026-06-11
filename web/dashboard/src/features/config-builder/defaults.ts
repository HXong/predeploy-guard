import type { ConfigBuilderState } from "./types";

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
  settings: {
    cleanup: true,
    timeoutSeconds: 60,
  },
};
