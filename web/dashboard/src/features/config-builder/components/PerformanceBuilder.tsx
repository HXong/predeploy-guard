import type { ChangeEvent, FormEvent } from "react";
import { createBlankPerformanceEndpoint, createHealthPerformanceEndpoint } from "../defaults";
import type { ConfigBuilderState, PerformanceConfigDraft, PerformanceEndpointDraft } from "../types";
import { PerformanceEndpointEditor } from "./PerformanceEndpointEditor";

type PerformanceBuilderProps = {
  config: ConfigBuilderState;
  onChange: (config: ConfigBuilderState) => void;
};

export function PerformanceBuilder({ config, onChange }: PerformanceBuilderProps) {
  const performance = config.performance;

  function updatePerformance(patch: Partial<PerformanceConfigDraft>) {
    onChange({
      ...config,
      performance: {
        ...performance,
        ...patch,
      },
    });
  }

  function updateThresholds(patch: Partial<PerformanceConfigDraft["thresholds"]>) {
    updatePerformance({
      thresholds: {
        ...performance.thresholds,
        ...patch,
      },
    });
  }

  function updateEndpoints(endpoints: PerformanceEndpointDraft[]) {
    updatePerformance({ endpoints });
  }

  function addEndpoint(endpoint: PerformanceEndpointDraft) {
    updateEndpoints([...performance.endpoints, endpoint]);
  }

  function updateEndpoint(id: string, patch: Partial<PerformanceEndpointDraft>) {
    updateEndpoints(
      performance.endpoints.map((endpoint) => (endpoint.id === id ? { ...endpoint, ...patch } : endpoint)),
    );
  }

  function removeEndpoint(id: string) {
    updateEndpoints(performance.endpoints.filter((endpoint) => endpoint.id !== id));
  }

  function handleEnabledChange(event: ChangeEvent<HTMLInputElement>) {
    updatePerformance({ enabled: event.target.checked });
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <form className="builder-form" onSubmit={handleSubmit}>
      <label className="checkbox-row">
        <input checked={performance.enabled} type="checkbox" onChange={handleEnabledChange} />
        Enable performance testing
      </label>

      {!performance.enabled ? (
        <p className="helper-text">Generated YAML will contain only performance enabled: false.</p>
      ) : (
        <>
          <div className="form-grid">
            <label>
              Virtual users
              <input
                min="1"
                step="1"
                type="number"
                value={Number.isFinite(performance.vus) ? performance.vus : ""}
                onChange={(event) => updatePerformance({ vus: numberInputValue(event.currentTarget) })}
              />
            </label>
            <label>
              Duration
              <input
                value={performance.duration}
                onChange={(event) => updatePerformance({ duration: event.target.value })}
                placeholder="15s"
              />
            </label>
            <label>
              Maximum p95 latency (ms)
              <input
                min="0.01"
                step="any"
                type="number"
                value={
                  Number.isFinite(performance.thresholds.maxP95LatencyMs)
                    ? performance.thresholds.maxP95LatencyMs
                    : ""
                }
                onChange={(event) =>
                  updateThresholds({ maxP95LatencyMs: numberInputValue(event.currentTarget) })
                }
              />
            </label>
            <label>
              Maximum error rate
              <input
                max="1"
                min="0.0001"
                step="0.001"
                type="number"
                value={
                  Number.isFinite(performance.thresholds.maxErrorRate)
                    ? performance.thresholds.maxErrorRate
                    : ""
                }
                onChange={(event) => updateThresholds({ maxErrorRate: numberInputValue(event.currentTarget) })}
              />
              <span className="helper-text">0.01 means 1% errors.</span>
            </label>
          </div>

          <div className="env-editor">
            <div className="section-header">
              <h3>Endpoints</h3>
              <div className="panel-actions">
                <button
                  className="secondary-button"
                  type="button"
                  onClick={() => addEndpoint(createBlankPerformanceEndpoint())}
                >
                  Add endpoint
                </button>
                <button
                  className="secondary-button"
                  type="button"
                  onClick={() => addEndpoint(createHealthPerformanceEndpoint())}
                >
                  Add health preset
                </button>
              </div>
            </div>

            {performance.endpoints.length === 0 ? (
              <p className="helper-text">
                {config.checks.smoke.length > 0
                  ? "No explicit performance endpoints. PreDeploy Guard will derive them from the configured smoke checks."
                  : "Add a performance endpoint or configure at least one smoke check."}
              </p>
            ) : (
              <div className="performance-endpoint-list">
                {performance.endpoints.map((endpoint) => (
                  <PerformanceEndpointEditor
                    endpoint={endpoint}
                    key={endpoint.id}
                    onChange={(patch) => updateEndpoint(endpoint.id, patch)}
                    onRemove={() => removeEndpoint(endpoint.id)}
                  />
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </form>
  );
}

function numberInputValue(input: HTMLInputElement): number {
  return input.value === "" ? Number.NaN : input.valueAsNumber;
}
