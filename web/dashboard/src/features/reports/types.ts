export type RuntimeEnvironmentReport = {
  runtime?: string;
  name?: string;
  baseUrl?: string;
  workloadBaseUrl?: string;
};

export type RunPhaseReport = {
  name: string;
  status: string;
  startedAt?: string;
  finishedAt?: string;
  duration?: number;
  error?: string;
};

export type BuildResultReport = {
  Enabled: boolean;
  Image: string;
  Context: string;
  Dockerfile: string;
  Passed: boolean;
  Error: string;
  Output: string;
};

export type ReadinessResultReport = {
  Name: string;
  Target: string;
  Passed: boolean;
  Error: string;
};

export type SmokeResultReport = {
  name: string;
  method: string;
  url: string;
  expectedStatus: number;
  actualStatus: number;
  passed: boolean;
  error?: string;
  duration: number;
  durationMs: number;
};

export type WorkloadResultReport = {
  Name: string;
  Type: string;
  Enabled: boolean;
  Passed: boolean;
  FailurePolicy: string;
  Method: string;
  Path: string;
  TargetURL: string;
  Duration: string;
  RatePerSecond: number;
  ExpectedStatus: number;
  StartedAt: string;
  FinishedAt: string;
  RequestCount: number;
  SuccessCount: number;
  FailureCount: number;
  StatusCounts: Record<string, number>;
  Error: string;
};

export type GatewayResultReport = {
  name: string;
  method: string;
  path: string;
  expectedStatus: number;
  gatewayUrl: string;
  gatewayStatus: number;
  gatewayDuration: number;
  gatewayPassed: boolean;
  gatewayError?: string;
  compareDirect: boolean;
  directUrl?: string;
  directStatus?: number;
  directDuration?: number;
  directPassed: boolean;
  directError?: string;
  statusMatched: boolean;
  gatewayLatencyMs: number;
  directLatencyMs?: number;
  latencyCompared: boolean;
  overheadMs?: number;
  overheadRatio?: number;
  latencyPassed: boolean;
  latencyWarnings?: string[];
  latencyErrors?: string[];
  passed: boolean;
};

export type PerformanceResultReport = {
  enabled: boolean;
  passed: boolean;
  vus: number;
  duration: string;
  avgLatencyMs: number;
  minLatencyMs: number;
  medianLatencyMs: number;
  maxLatencyMs: number;
  p90LatencyMs: number;
  p95LatencyMs: number;
  maxP95LatencyMs: number;
  errorRate: number;
  maxErrorRate: number;
  requestCount: number;
  iterations: number;
  checksTotal: number;
  checkPassRate: number;
  error?: string;
  output?: string;
};

export type RunReport = {
  ServiceName?: string;
  activeProfile?: string;
  Image?: string;
  Runtime?: string;
  BaseURL?: string;
  runtimeEnvironment?: RuntimeEnvironmentReport;
  runPhases?: RunPhaseReport[];
  StartedAt?: string;
  FinishedAt?: string;
  BuildResult?: BuildResultReport;
  ReadinessResults?: ReadinessResultReport[];
  GatewayResults?: GatewayResultReport[];
  Results?: SmokeResultReport[];
  WorkloadResults?: WorkloadResultReport[];
  PerformanceResult?: PerformanceResultReport;
  Passed?: boolean;
  Logs?: string;
};
