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
