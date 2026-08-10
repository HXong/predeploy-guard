export type GatewayIngressSummary = {
  enabled: boolean;
  host?: string;
  className?: string;
  pathType?: string;
  annotationCount: number;
};

export type GatewayLatencySummary = {
  enabled: boolean;
  maxGatewayLatencyMs?: number;
  maxOverheadMs?: number;
  maxOverheadRatio?: number;
  failurePolicy?: string;
};

export type GatewayRouteSummary = {
  name: string;
  method: string;
  path: string;
  expectedStatus: number;
  compareDirect: boolean;
};

export type GatewaySummary = {
  enabled: boolean;
  baseUrl?: string;
  ingress: GatewayIngressSummary;
  latency: GatewayLatencySummary;
  routes: GatewayRouteSummary[];
};

export type WorkloadSummary = {
  name: string;
  type: string;
  enabled: boolean;
  method: string;
  path: string;
  duration: string;
  ratePerSecond: number;
  expectedStatus: number;
  failurePolicy: string;
};

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
  gateway: GatewaySummary;
  smokeChecks: Array<{
    name: string;
    method: string;
    path: string;
    expectedStatus: number;
  }>;
  workloads: WorkloadSummary[];
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
    gatewayRoutes: number;
    workloads: number;
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
