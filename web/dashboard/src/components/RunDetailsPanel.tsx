import type { RunHistoryItem } from "../types/api";
import {
  formatDate,
  formatDuration,
  formatErrorRate,
  formatLatency,
  formatProfile,
  formatResult,
  shortId,
} from "../utils/format";
import { getReportPath } from "../utils/reports";
import { Card } from "./Card";

type MarkdownPreview = {
  data: string | null;
  error: string | null;
  loading: boolean;
};

type RunDetailsPanelProps = {
  markdownPreview: MarkdownPreview;
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

function Detail({ label, value, title }: { label: string; value: string; title?: string }) {
  return (
    <div className="detail-item">
      <span>{label}</span>
      <strong title={title}>{value}</strong>
    </div>
  );
}
