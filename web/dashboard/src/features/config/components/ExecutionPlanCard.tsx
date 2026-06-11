import { Card } from "../../../shared/components";
import type { ConfigExplanation } from "../types";

type ExecutionPlanCardProps = {
  data: ConfigExplanation | null;
  error: string | null;
  loading: boolean;
};

export function ExecutionPlanCard({ data, error, loading }: ExecutionPlanCardProps) {
  return (
    <Card title={data?.title ?? "Execution Plan"} loading={loading} error={error}>
      <div className="plan-list">
        {(data?.steps.slice(0, 4) ?? []).map((step) => (
          <div className="plan-step" key={step.number}>
            <span>{step.number}</span>
            <div>
              <strong>{step.title}</strong>
              <p>{step.details[0] ?? "No additional details."}</p>
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}
