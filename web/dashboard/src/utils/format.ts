export function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function formatDuration(startedAt: string, finishedAt: string): string {
  const started = new Date(startedAt);
  const finished = new Date(finishedAt);
  const milliseconds = finished.getTime() - started.getTime();

  if (Number.isNaN(milliseconds) || milliseconds < 0) {
    return "Unknown";
  }

  if (milliseconds < 1000) {
    return `${milliseconds} ms`;
  }

  const totalSeconds = Math.round(milliseconds / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`;
  }

  if (minutes > 0) {
    return `${minutes}m ${seconds}s`;
  }

  return `${seconds}s`;
}

export function formatProfile(profile?: string): string {
  return profile || "base";
}

export function formatResult(passed: boolean): string {
  return passed ? "Passed" : "Failed";
}

export function formatLatency(value: number): string {
  return `${value.toFixed(2)} ms`;
}

export function formatErrorRate(value: number): string {
  return `${(value * 100).toFixed(2)}%`;
}

export function shortId(value: string): string {
  if (value.length <= 16) {
    return value;
  }

  return `${value.slice(0, 10)}...${value.slice(-4)}`;
}

export function isSelectedProfile(candidate: string, value: string): boolean {
  const normalized = value.trim();
  return candidate === "base" ? normalized === "" || normalized === "base" : normalized === candidate;
}
