import type { ReportType } from "../types/api";

export function getReportPath(runId: string, reportType: ReportType): string {
  return `/api/reports/${encodeURIComponent(runId)}/${reportType}`;
}
