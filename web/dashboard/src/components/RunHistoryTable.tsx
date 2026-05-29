import type { RunHistoryItem } from "../types/api";
import { formatDate, formatProfile, formatResult, shortId } from "../utils/format";
import { getReportPath } from "../utils/reports";
import { Card } from "./Card";

type RunHistoryTableProps = {
  error: string | null;
  loading: boolean;
  onSelect: (runId: string) => void;
  runs: RunHistoryItem[];
  selectedRunId: string;
};

export function RunHistoryTable({ error, loading, onSelect, runs, selectedRunId }: RunHistoryTableProps) {
  return (
    <Card title="Run History" loading={loading} error={error}>
      {runs.length === 0 ? (
        <p className="empty-state">No completed run history yet.</p>
      ) : (
        <div className="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Run</th>
                <th>Service</th>
                <th>Profile</th>
                <th>Result</th>
                <th>Finished</th>
                <th>Reports</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((run) => (
                <tr
                  className={run.runId === selectedRunId ? "selected-row" : ""}
                  key={run.runId}
                  onClick={() => onSelect(run.runId)}
                >
                  <td title={run.runId}>{shortId(run.runId)}</td>
                  <td>{run.serviceName}</td>
                  <td>{formatProfile(run.profile)}</td>
                  <td>{formatResult(run.passed)}</td>
                  <td>{formatDate(run.finishedAt)}</td>
                  <td>
                    <div className="report-links" onClick={(event) => event.stopPropagation()}>
                      <a href={getReportPath(run.runId, "markdown")} target="_blank" rel="noreferrer">
                        Markdown
                      </a>
                      <a href={getReportPath(run.runId, "json")} target="_blank" rel="noreferrer">
                        JSON
                      </a>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  );
}
