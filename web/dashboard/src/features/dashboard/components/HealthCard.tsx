import { Card } from "../../../shared/components";
import type { HealthResponse } from "../types";

type HealthCardProps = {
  data: HealthResponse | null;
  error: string | null;
  loading: boolean;
};

export function HealthCard({ data, error, loading }: HealthCardProps) {
  return (
    <Card title="API Health" loading={loading} error={error}>
      {data && (
        <div className="status-row">
          <span className={`status-dot ${data.status === "ok" ? "ok" : "warn"}`} />
          <span className="status-text">{data.status}</span>
        </div>
      )}
    </Card>
  );
}
