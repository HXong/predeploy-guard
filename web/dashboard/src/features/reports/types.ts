/** A Go time.Duration encoded as its raw JSON nanosecond count. */
export type GoDurationNanoseconds = number;

/** A backend value explicitly normalized to milliseconds. */
export type Milliseconds = number;

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
  duration?: GoDurationNanoseconds;
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
  duration: GoDurationNanoseconds;
  durationMs: Milliseconds;
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
  gatewayDuration: GoDurationNanoseconds;
  gatewayPassed: boolean;
  gatewayError?: string;
  compareDirect: boolean;
  directUrl?: string;
  directStatus?: number;
  directDuration?: GoDurationNanoseconds;
  directPassed: boolean;
  directError?: string;
  statusMatched: boolean;
  gatewayLatencyMs: Milliseconds;
  directLatencyMs?: Milliseconds;
  latencyCompared: boolean;
  overheadMs?: Milliseconds;
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
  avgLatencyMs: Milliseconds;
  minLatencyMs: Milliseconds;
  medianLatencyMs: Milliseconds;
  maxLatencyMs: Milliseconds;
  p90LatencyMs: Milliseconds;
  p95LatencyMs: Milliseconds;
  maxP95LatencyMs: Milliseconds;
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
