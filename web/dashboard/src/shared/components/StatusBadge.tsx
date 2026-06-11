type StatusBadgeProps = {
  status: "queued" | "running" | "completed" | "failed" | string;
};

export function StatusBadge({ status }: StatusBadgeProps) {
  return <span className={`status-badge ${status}`}>{status}</span>;
}
