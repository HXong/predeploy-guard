import { request, requestText } from "../../shared/api/http";
import type { RunReport } from "./types";

export function getJSONReport(runId: string): Promise<RunReport> {
  return request<RunReport>(`/reports/${encodeURIComponent(runId)}/json`);
}

export function getMarkdownReport(runId: string): Promise<string> {
  return requestText(`/reports/${encodeURIComponent(runId)}/markdown`);
}
