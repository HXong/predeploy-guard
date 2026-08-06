import type { GoDurationNanoseconds } from "./types";

export type ReportType = "markdown" | "json";

const nanosecondsPerMillisecond = 1_000_000;
const millisecondsPerSecond = 1_000;
const secondsPerMinute = 60;

export function formatReportDuration(value: GoDurationNanoseconds | null | undefined): string {
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0) {
    return "-";
  }

  if (value === 0) {
    return "0ms";
  }

  const milliseconds = value / nanosecondsPerMillisecond;
  if (milliseconds < 1) {
    return "< 1ms";
  }
  if (milliseconds < millisecondsPerSecond) {
    const displayedMilliseconds = milliseconds < 10 ? compactDecimal(milliseconds) : Math.round(milliseconds);
    return `${displayedMilliseconds}ms`;
  }

  const seconds = milliseconds / millisecondsPerSecond;
  if (seconds < secondsPerMinute) {
    const displayedSeconds = seconds < 10 ? compactDecimal(seconds) : Math.round(seconds);
    return `${displayedSeconds}s`;
  }

  const totalSeconds = Math.round(seconds);
  const minutes = Math.floor(totalSeconds / secondsPerMinute);
  const remainingSeconds = totalSeconds % secondsPerMinute;
  return `${minutes}m ${String(remainingSeconds).padStart(2, "0")}s`;
}

export function formatReportValue(value: string | null | undefined): string {
  const normalizedValue = value?.trim();
  return normalizedValue || "-";
}

export function getReportPath(runId: string, reportType: ReportType): string {
  return `/api/reports/${encodeURIComponent(runId)}/${reportType}`;
}

function compactDecimal(value: number): string {
  return Number(value.toFixed(1)).toString();
}
