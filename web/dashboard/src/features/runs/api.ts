import { request } from "../../shared/api/http";
import type { RunHistoryItem, RunTask, RunTaskLogsResponse } from "./types";

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
