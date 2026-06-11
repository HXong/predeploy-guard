import { Card, StatusBadge } from "../../../shared/components";
import type { RunTask } from "../types";

type TaskLogsPanelProps = {
  error: string | null;
  loading: boolean;
  logs: string[];
  task: RunTask | null;
  taskError: string | null;
};

export function TaskLogsPanel({ error, loading, logs, task, taskError }: TaskLogsPanelProps) {
  return (
    <Card title="Task Logs" loading={loading} error={error}>
      <TaskLogs task={task} logs={logs} taskError={taskError} />
    </Card>
  );
}

function TaskLogs({ task, logs, taskError }: Omit<TaskLogsPanelProps, "error" | "loading">) {
  if (taskError) {
    return <p className="error-text">{taskError}</p>;
  }

  if (!task) {
    return <p className="empty-state">Select a task to inspect its logs.</p>;
  }

  return (
    <div className="logs-panel">
      <div className="task-summary">
        <StatusBadge status={task.status} />
        <span>{task.message || task.error || task.taskId}</span>
      </div>
      <pre>{logs.length > 0 ? logs.join("\n") : "No logs recorded yet."}</pre>
    </div>
  );
}
