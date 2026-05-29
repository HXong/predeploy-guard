import type { RunTask } from "../types/api";

type StatusBadgeProps = {
  status: RunTask["status"];
};

export function StatusBadge({ status }: StatusBadgeProps) {
  return <span className={`status-badge ${status}`}>{status}</span>;
}
