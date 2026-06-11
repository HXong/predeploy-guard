export type ConfigSummary = {
  configPath: string;
  configDir: string;
  activeProfile?: string;
  availableProfiles: string[];
  runtime: {
    type: string;
  };
  service: {
    name: string;
    image: string;
    port: number;
    healthPath: string;
    envCount: number;
  };
  build: {
    enabled: boolean;
    context?: string;
    dockerfile?: string;
  };
  dependencies: Array<{
    name: string;
    image: string;
    port?: number;
    envCount: number;
    hasReadiness: boolean;
    readiness: string;
  }>;
  smokeChecks: Array<{
    name: string;
    method: string;
    path: string;
    expectedStatus: number;
  }>;
  performance: {
    enabled: boolean;
    vus?: number;
    duration?: string;
    maxP95LatencyMs?: number;
    maxErrorRate?: number;
    endpointCount: number;
  };
  settings: {
    namespacePrefix: string;
    cleanup: boolean;
    timeoutSeconds: number;
  };
  reports: {
    directory: string;
    markdown: boolean;
    json: boolean;
    history: boolean;
  };
  counts: {
    dependencies: number;
    dependenciesWithReadiness: number;
    smokeChecks: number;
    performanceEndpoints: number;
    profiles: number;
  };
};

export type ConfigExplanation = {
  title: string;
  steps: Array<{
    number: number;
    title: string;
    details: string[];
  }>;
};
