import { Card } from "../../../shared/components";
import type { ConfigSummary } from "../types";

type ConfigSummaryCardProps = {
  data: ConfigSummary | null;
  error: string | null;
  loading: boolean;
};

export function ConfigSummaryCard({ data, error, loading }: ConfigSummaryCardProps) {
  return (
    <Card title="Config Summary" loading={loading} error={error}>
      {data && (
        <div className="summary-grid config-summary-grid">
          <Metric label="Service" value={formatSummaryText(data.service?.name)} />
          <Metric label="Runtime" value={formatSummaryText(data.runtime?.type)} />
          <Metric label="Port" value={formatSummaryNumber(data.service?.port)} />
          <Metric label="Profiles" value={formatSummaryCount(data.counts?.profiles, "profile")} />
          <Metric label="Smoke checks" value={formatSummaryCount(data.counts?.smokeChecks, "check")} />
          <Metric label="Dependencies" value={formatSummaryCount(data.counts?.dependencies, "dependency")} />
          <Metric label="Gateway routes" value={formatSummaryCount(data.counts?.gatewayRoutes, "route")} />
          <Metric label="Workloads" value={formatSummaryCount(data.counts?.workloads, "workload")} />
          <Metric label="Gateway latency" value={formatLatencySummary(data)} />
          <Metric label="Ingress" value={formatEnabledLabel(data.gateway?.ingress?.enabled)} />
          <Metric label="Performance" value={formatPerformanceSummary(data)} />
        </div>
      )}
    </Card>
  );
}

type MetricProps = {
  label: string;
  value: string;
};

function Metric({ label, value }: MetricProps) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function formatSummaryText(value: string | null | undefined): string {
  const normalizedValue = value?.trim();
  return normalizedValue || "-";
}

function formatSummaryNumber(value: number | null | undefined): string {
  return typeof value === "number" && Number.isFinite(value) ? String(value) : "-";
}

function formatSummaryCount(value: number | null | undefined, singularLabel: string): string {
  if (typeof value !== "number" || !Number.isFinite(value) || value < 0) {
    return "-";
  }

  return `${value} ${value === 1 ? singularLabel : pluralizeSummaryLabel(singularLabel)}`;
}

function formatEnabledLabel(value: boolean | null | undefined): string {
  if (typeof value !== "boolean") {
    return "-";
  }

  return value ? "Enabled" : "Disabled";
}

function pluralizeSummaryLabel(label: string): string {
  return label.endsWith("y") ? `${label.slice(0, -1)}ies` : `${label}s`;
}

function formatLatencySummary(data: ConfigSummary): string {
  const enabled = data.gateway?.latency?.enabled;
  if (enabled !== true) {
    return formatEnabledLabel(enabled);
  }

  const policy = formatSummaryText(data.gateway.latency.failurePolicy);
  return policy === "-" ? "Enabled" : `Enabled · ${policy}`;
}

function formatPerformanceSummary(data: ConfigSummary): string {
  const enabled = data.performance?.enabled;
  if (enabled !== true) {
    return formatEnabledLabel(enabled);
  }

  const endpoints = formatSummaryCount(data.counts?.performanceEndpoints, "endpoint");
  return endpoints === "-" ? "Enabled" : `Enabled · ${endpoints}`;
}
