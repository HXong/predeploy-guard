import { Card } from "../../../shared/components";
import type { Resource } from "../../../shared/utils";
import {
  formatDate,
  formatDuration,
  formatErrorRate,
  formatLatency,
  formatProfile,
  formatResult,
  shortId,
} from "../../../shared/utils";
import type { RunHistoryItem } from "../../runs";
import { getReportPath } from "../utils";

type RunDetailsPanelProps = {
  markdownPreview: Resource<string>;
  onPreviewMarkdown: () => void;
  run: RunHistoryItem | null;
};

export function RunDetailsPanel({ markdownPreview, onPreviewMarkdown, run }: RunDetailsPanelProps) {
  return (
    <Card title="Run Details">
      {!run ? (
        <p className="empty-state">Select a run from history to inspect its details.</p>
      ) : (
        <div className="run-details">
          <div className="details-grid">
            <Detail label="Run ID" title={run.runId} value={shortId(run.runId)} />
            <Detail label="Service" value={run.serviceName} />
            <Detail label="Profile" value={formatProfile(run.profile)} />
            <Detail label="Result" value={formatResult(run.passed)} />
            <Detail label="Started" value={formatDate(run.startedAt)} />
            <Detail label="Finished" value={formatDate(run.finishedAt)} />
            <Detail label="Duration" value={formatDuration(run.startedAt, run.finishedAt)} />
            {run.p95LatencyMs !== undefined && <Detail label="p95 latency" value={formatLatency(run.p95LatencyMs)} />}
            {run.errorRate !== undefined && <Detail label="Error rate" value={formatErrorRate(run.errorRate)} />}
          </div>

          <div className="panel-actions">
            <a href={getReportPath(run.runId, "markdown")} target="_blank" rel="noreferrer">
              Markdown report
            </a>
            <a href={getReportPath(run.runId, "json")} target="_blank" rel="noreferrer">
              JSON report
            </a>
            <button
              className="secondary-button"
              disabled={markdownPreview.loading}
              type="button"
              onClick={onPreviewMarkdown}
            >
              {markdownPreview.loading ? "Loading preview..." : "Preview Markdown"}
            </button>
          </div>

          {markdownPreview.error && <p className="error-text">{markdownPreview.error}</p>}
          {markdownPreview.data !== null && <pre className="markdown-preview">{markdownPreview.data}</pre>}
        </div>
      )}
    </Card>
  );
}

type DetailProps = {
  label: string;
  value: string;
  title?: string;
};

function Detail({ label, value, title }: DetailProps) {
  return (
    <div className="detail-item">
      <span>{label}</span>
      <strong title={title}>{value}</strong>
    </div>
  );
}
