export type HealthResponse = {
  status: string;
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

export type RunHistoryItem = {
  runId: string;
  serviceName: string;
  profile?: string;
  passed: boolean;
  startedAt: string;
  finishedAt: string;
  markdownPath: string;
  jsonPath: string;
  p95LatencyMs?: number;
  errorRate?: number;
};

export type RunTaskStatus = "queued" | "running" | "completed" | "failed";

export type RunTask = {
  taskId: string;
  profile: string;
  status: RunTaskStatus;
  passed: boolean;
  error?: string;
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
  message?: string;
  logs: string[];
};

export type RunTaskLogsResponse = {
  taskId: string;
  logs: string[];
};

export type ReportType = "markdown" | "json";

type ApiErrorBody = {
  error?: string;
  message?: string;
};

const configuredBaseUrl = import.meta.env.VITE_PREDEPLOY_API_BASE_URL?.replace(/\/+$/, "") ?? "";

function buildApiUrl(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  if (!configuredBaseUrl) {
    return `/api${normalizedPath}`;
  }

  const baseIncludesApi = configuredBaseUrl.endsWith("/api");
  return `${configuredBaseUrl}${baseIncludesApi ? "" : "/api"}${normalizedPath}`;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;

  try {
    response = await fetch(buildApiUrl(path), {
      headers: {
        Accept: "application/json",
        ...init?.headers,
      },
      ...init,
    });
  } catch (error) {
    throw new Error("Could not reach the PreDeploy Guard API. Start the backend and try again.");
  }

  if (!response.ok) {
    const message = await readErrorMessage(response);
    throw new Error(message || `API request failed with HTTP ${response.status}.`);
  }

  return response.json() as Promise<T>;
}

async function requestText(path: string): Promise<string> {
  let response: Response;

  try {
    response = await fetch(buildApiUrl(path), {
      headers: {
        Accept: "text/markdown, text/plain, */*",
      },
    });
  } catch (error) {
    throw new Error("Could not reach the PreDeploy Guard API. Start the backend and try again.");
  }

  if (!response.ok) {
    const message = await readTextErrorMessage(response);
    throw new Error(message || `API request failed with HTTP ${response.status}.`);
  }

  return response.text();
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as ApiErrorBody;
    return body.error ?? body.message ?? "";
  } catch {
    return "";
  }
}

async function readTextErrorMessage(response: Response): Promise<string> {
  const contentType = response.headers.get("Content-Type") ?? "";
  if (contentType.includes("application/json")) {
    return readErrorMessage(response);
  }

  try {
    return await response.text();
  } catch {
    return "";
  }
}

export function getReportPath(runId: string, reportType: ReportType): string {
  return `/api/reports/${encodeURIComponent(runId)}/${reportType}`;
}

export function getHealth(): Promise<HealthResponse> {
  return request<HealthResponse>("/health");
}

export function getConfigSummary(): Promise<ConfigSummary> {
  return request<ConfigSummary>("/config/summary");
}

export function getConfigExplain(): Promise<ConfigExplanation> {
  return request<ConfigExplanation>("/config/explain");
}

export function getHistory(): Promise<RunHistoryItem[]> {
  return request<RunHistoryItem[]>("/history");
}

export function getRunTasks(): Promise<RunTask[]> {
  return request<RunTask[]>("/runs");
}

export function triggerRun(profile: string): Promise<RunTask> {
  const normalizedProfile = profile.trim();
  const requestProfile = normalizedProfile === "base" ? "" : normalizedProfile;

  return request<RunTask>("/runs", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ profile: requestProfile }),
  });
}

export function getRunTask(taskId: string): Promise<RunTask> {
  return request<RunTask>(`/runs/${encodeURIComponent(taskId)}`);
}

export function getRunTaskLogs(taskId: string): Promise<RunTaskLogsResponse> {
  return request<RunTaskLogsResponse>(`/runs/${encodeURIComponent(taskId)}/logs`);
}

export function getMarkdownReport(runId: string): Promise<string> {
  return requestText(`/reports/${encodeURIComponent(runId)}/markdown`);
}
