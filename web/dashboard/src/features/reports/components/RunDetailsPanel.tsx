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
import type { RunReport, RuntimeEnvironmentReport } from "../types";
import { formatReportDuration, formatReportValue, getReportPath } from "../utils";

type RunDetailsPanelProps = {
  markdownPreview: Resource<string>;
  onPreviewMarkdown: () => void;
  report: Resource<RunReport>;
  run: RunHistoryItem | null;
};

export function RunDetailsPanel({ markdownPreview, onPreviewMarkdown, report, run }: RunDetailsPanelProps) {
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

          <StructuredReportDetails report={report} />

          {markdownPreview.error && <p className="error-text">{markdownPreview.error}</p>}
          {markdownPreview.data !== null && <pre className="markdown-preview">{markdownPreview.data}</pre>}
        </div>
      )}
    </Card>
  );
}

type StructuredReportDetailsProps = {
  report: Resource<RunReport>;
};

function StructuredReportDetails({ report }: StructuredReportDetailsProps) {
  if (report.loading) {
    return <p className="empty-state structured-report-state">Loading structured report details...</p>;
  }

  if (report.error) {
    return (
      <div className="structured-report-state">
        <p className="empty-state">Structured report details are unavailable.</p>
        <p className="error-text">{report.error}</p>
      </div>
    );
  }

  if (!report.data) {
    return null;
  }

  return (
    <div className="structured-report">
      <RuntimeEnvironmentSection environment={report.data.runtimeEnvironment} />
      <RunTimelineSection phases={report.data.runPhases} />
    </div>
  );
}

type RuntimeEnvironmentSectionProps = {
  environment?: RuntimeEnvironmentReport;
};

function RuntimeEnvironmentSection({ environment }: RuntimeEnvironmentSectionProps) {
  return (
    <section className="report-section">
      <h3>Runtime Environment</h3>
      {!environment ? (
        <p className="empty-state">Runtime environment details were not recorded for this run.</p>
      ) : (
        <div className="details-grid report-details-grid">
          <Detail label="Runtime" value={formatReportValue(environment.runtime)} />
          <Detail label="Environment" value={formatReportValue(environment.name)} />
          <Detail label="Base URL" value={formatReportValue(environment.baseUrl)} />
          <Detail label="Workload Base URL" value={formatReportValue(environment.workloadBaseUrl)} />
        </div>
      )}
    </section>
  );
}

type RunTimelineSectionProps = {
  phases?: RunReport["runPhases"];
};

function RunTimelineSection({ phases }: RunTimelineSectionProps) {
  return (
    <section className="report-section">
      <h3>Run Timeline</h3>
      {!phases?.length ? (
        <p className="empty-state">Run phase timings were not recorded for this run.</p>
      ) : (
        <div className="table-wrap">
          <table className="run-timeline-table">
            <thead>
              <tr>
                <th scope="col">Phase</th>
                <th scope="col">Status</th>
                <th scope="col">Duration</th>
                <th scope="col">Details</th>
              </tr>
            </thead>
            <tbody>
              {phases.map((phase, index) => (
                <tr key={`${phase.name}-${index}`}>
                  <td>{formatReportValue(phase.name)}</td>
                  <td>{formatReportValue(phase.status)}</td>
                  <td>{formatReportDuration(phase.duration)}</td>
                  <td>{formatReportValue(phase.error)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
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
