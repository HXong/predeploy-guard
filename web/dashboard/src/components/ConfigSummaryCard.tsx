import type { ConfigSummary } from "../types/api";
import { Card } from "./Card";

type ConfigSummaryCardProps = {
  data: ConfigSummary | null;
  error: string | null;
  loading: boolean;
};

export function ConfigSummaryCard({ data, error, loading }: ConfigSummaryCardProps) {
  return (
    <Card title="Config Summary" loading={loading} error={error}>
      {data && (
        <div className="summary-grid">
          <Metric label="Service" value={data.service.name} />
          <Metric label="Runtime" value={data.runtime.type} />
          <Metric label="Port" value={String(data.service.port)} />
          <Metric label="Profiles" value={String(data.counts.profiles)} />
          <Metric label="Smoke checks" value={String(data.counts.smokeChecks)} />
          <Metric label="Dependencies" value={String(data.counts.dependencies)} />
        </div>
      )}
    </Card>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
