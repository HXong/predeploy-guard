import { Card, StatusBadge } from "../../../shared/components";
import { formatDate, formatProfile, shortId } from "../../../shared/utils";
import type { RunTask } from "../types";

type RunTasksTableProps = {
  error: string | null;
  loading: boolean;
  onSelect: (taskId: string) => void;
  selectedTaskId: string;
  tasks: RunTask[];
};

export function RunTasksTable({ error, loading, onSelect, selectedTaskId, tasks }: RunTasksTableProps) {
  return (
    <Card title="Run Tasks" loading={loading} error={error}>
      {tasks.length === 0 ? (
        <p className="empty-state">No API-triggered runs yet.</p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Task</th>
                <th>Profile</th>
                <th>Status</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              {tasks.map((task) => (
                <tr
                  className={task.taskId === selectedTaskId ? "selected-row" : ""}
                  key={task.taskId}
                  onClick={() => onSelect(task.taskId)}
                >
                  <td title={task.taskId}>{shortId(task.taskId)}</td>
                  <td>{formatProfile(task.profile)}</td>
                  <td>
                    <StatusBadge status={task.status} />
                  </td>
                  <td>{formatDate(task.createdAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  );
}
