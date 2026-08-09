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
import type { GatewayResultReport, RunReport, RuntimeEnvironmentReport } from "../types";
import {
  formatBooleanResult,
  formatMilliseconds,
  formatRatio,
  formatReportDuration,
  formatReportValue,
  formatStatusCode,
  formatWarningList,
  getReportPath,
} from "../utils";

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
      <GatewayResultsSection results={report.data.GatewayResults} />
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

type GatewayResultsSectionProps = {
  results?: GatewayResultReport[];
};

function GatewayResultsSection({ results }: GatewayResultsSectionProps) {
  return (
    <section className="report-section">
      <h3>Gateway Results</h3>
      {!results?.length ? (
        <p className="empty-state">Gateway results were not recorded for this run.</p>
      ) : (
        <div className="table-wrap">
          <table className="gateway-results-table">
            <thead>
              <tr>
                <th scope="col">Route</th>
                <th scope="col">Expected</th>
                <th scope="col">Gateway</th>
                <th scope="col">Direct</th>
                <th scope="col">Match</th>
                <th scope="col">Latency</th>
                <th scope="col">Overhead</th>
                <th scope="col">Result</th>
                <th scope="col">Details</th>
              </tr>
            </thead>
            <tbody>
              {results.map((result, index) => (
                <GatewayResultRow key={`${result.name}-${result.method}-${result.path}-${index}`} result={result} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

type GatewayResultRowProps = {
  result: GatewayResultReport;
};

function GatewayResultRow({ result }: GatewayResultRowProps) {
  const details = gatewayResultDetails(result);
  const latencyResult = gatewayLatencyResult(result);
  const routeName = result.name?.trim();

  return (
    <tr>
      <td>
        <div className="report-cell-stack route-result-cell">
          {routeName && <strong>{routeName}</strong>}
          <span>{formatRoute(result.method, result.path)}</span>
        </div>
      </td>
      <td>{formatStatusCode(result.expectedStatus)}</td>
      <td>
        <StatusResult status={result.gatewayStatus} passed={result.gatewayPassed} />
      </td>
      <td>
        {result.compareDirect ? (
          <StatusResult status={result.directStatus} passed={result.directPassed} />
        ) : (
          <span className="report-muted-value">Not compared</span>
        )}
      </td>
      <td>
        {result.compareDirect ? formatBooleanResult(result.statusMatched, "YES", "NO") : "-"}
      </td>
      <td>
        <div className="report-cell-stack">
          <ReportMetric label="Gateway latency" value={formatMilliseconds(result.gatewayLatencyMs)} />
          {result.directLatencyMs !== undefined && (
            <ReportMetric label="Direct latency" value={formatMilliseconds(result.directLatencyMs)} />
          )}
        </div>
      </td>
      <td>
        {result.overheadMs === undefined && result.overheadRatio === undefined ? (
          "-"
        ) : (
          <div className="report-cell-stack">
            <ReportMetric label="Overhead" value={formatMilliseconds(result.overheadMs)} />
            <ReportMetric label="Ratio" value={formatRatio(result.overheadRatio)} />
          </div>
        )}
      </td>
      <td>
        <div className="report-cell-stack">
          <ReportResult label="Latency" result={latencyResult.label} tone={latencyResult.tone} />
          <ReportResult
            label="Overall"
            result={formatBooleanResult(result.passed)}
            tone={result.passed ? "pass" : "fail"}
          />
        </div>
      </td>
      <td>
        {details.length ? (
          <ul className="report-message-list">
            {details.map((detail, index) => (
              <li key={`${detail}-${index}`}>{detail}</li>
            ))}
          </ul>
        ) : (
          "-"
        )}
      </td>
    </tr>
  );
}

type StatusResultProps = {
  passed: boolean;
  status?: number;
};

function StatusResult({ passed, status }: StatusResultProps) {
  return (
    <div className="report-cell-stack">
      <span>{formatStatusCode(status)}</span>
      <span className={passed ? "result-pass" : "result-fail"}>{formatBooleanResult(passed)}</span>
    </div>
  );
}

type ReportMetricProps = {
  label: string;
  value: string;
};

function ReportMetric({ label, value }: ReportMetricProps) {
  return (
    <span>
      <span className="report-cell-label">{label}:</span> {value}
    </span>
  );
}

type ResultTone = "pass" | "fail" | "warn" | "neutral";

type ReportResultProps = {
  label: string;
  result: string;
  tone: ResultTone;
};

function ReportResult({ label, result, tone }: ReportResultProps) {
  return (
    <span>
      <span className="report-cell-label">{label}:</span>{" "}
      <span className={`result-${tone}`}>{result}</span>
    </span>
  );
}

function formatRoute(method: string | null | undefined, path: string | null | undefined): string {
  const route = [method?.trim(), path?.trim()].filter((value): value is string => Boolean(value)).join(" ");
  return route || "-";
}

function gatewayResultDetails(result: GatewayResultReport): string[] {
  const details: string[] = [];
  const gatewayError = result.gatewayError?.trim();
  const directError = result.directError?.trim();

  if (gatewayError) {
    details.push(`Gateway: ${gatewayError}`);
  }
  if (result.compareDirect && directError) {
    details.push(`Direct service: ${directError}`);
  }

  details.push(...formatWarningList(result.latencyWarnings).map((warning) => `Latency warning: ${warning}`));
  details.push(...formatWarningList(result.latencyErrors).map((error) => `Latency error: ${error}`));
  return details;
}

function gatewayLatencyResult(result: GatewayResultReport): { label: string; tone: ResultTone } {
  const latencyErrors = formatWarningList(result.latencyErrors);
  const latencyWarnings = formatWarningList(result.latencyWarnings);

  if (!result.latencyCompared && latencyWarnings.length === 0 && latencyErrors.length === 0) {
    return { label: "Not compared", tone: "neutral" };
  }
  if (latencyErrors.length > 0) {
    return { label: "FAIL", tone: "fail" };
  }
  if (latencyWarnings.length > 0) {
    return { label: "WARN", tone: "warn" };
  }

  return {
    label: formatBooleanResult(result.latencyPassed),
    tone: result.latencyPassed ? "pass" : "fail",
  };
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
