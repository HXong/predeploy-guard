import type { ChangeEvent, FormEvent } from "react";
import { createBlankSmokeCheck, createHealthCheck } from "../defaults";
import type { ConfigBuilderState, HttpMethod, SmokeCheckDraft } from "../types";

const httpMethods: HttpMethod[] = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"];

type SmokeCheckBuilderProps = {
  config: ConfigBuilderState;
  onChange: (config: ConfigBuilderState) => void;
};

export function SmokeCheckBuilder({ config, onChange }: SmokeCheckBuilderProps) {
  function updateSmokeChecks(smoke: SmokeCheckDraft[]) {
    onChange({
      ...config,
      checks: {
        ...config.checks,
        smoke,
      },
    });
  }

  function addSmokeCheck(smokeCheck: SmokeCheckDraft) {
    updateSmokeChecks([...config.checks.smoke, smokeCheck]);
  }

  function removeSmokeCheck(id: string) {
    updateSmokeChecks(config.checks.smoke.filter((smokeCheck) => smokeCheck.id !== id));
  }

  function updateSmokeCheck(id: string, patch: Partial<SmokeCheckDraft>) {
    updateSmokeChecks(
      config.checks.smoke.map((smokeCheck) =>
        smokeCheck.id === id ? { ...smokeCheck, ...patch } : smokeCheck,
      ),
    );
  }

  function handleMethodChange(smokeCheckId: string, event: ChangeEvent<HTMLSelectElement>) {
    const method = event.target.value;
    if (isHttpMethod(method)) {
      updateSmokeCheck(smokeCheckId, { method });
    }
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <form className="builder-form" onSubmit={handleSubmit}>
      <div className="panel-actions">
        <button className="secondary-button" type="button" onClick={() => addSmokeCheck(createBlankSmokeCheck())}>
          Add smoke check
        </button>
        <button className="secondary-button" type="button" onClick={() => addSmokeCheck(createHealthCheck())}>
          Add health-check preset
        </button>
      </div>

      {config.checks.smoke.length === 0 ? (
        <p className="empty-state">No smoke checks configured.</p>
      ) : (
        <div className="smoke-check-list">
          {config.checks.smoke.map((smokeCheck) => (
            <section className="smoke-check-card" key={smokeCheck.id}>
              <div className="section-header">
                <h3>{smokeCheck.name || "Smoke check"}</h3>
                <button className="secondary-button" type="button" onClick={() => removeSmokeCheck(smokeCheck.id)}>
                  Remove
                </button>
              </div>

              <div className="form-grid">
                <label>
                  Name
                  <input
                    value={smokeCheck.name}
                    onChange={(event) => updateSmokeCheck(smokeCheck.id, { name: event.target.value })}
                  />
                </label>
                <label>
                  Method
                  <select
                    value={smokeCheck.method}
                    onChange={(event) => handleMethodChange(smokeCheck.id, event)}
                  >
                    {httpMethods.map((method) => (
                      <option key={method} value={method}>
                        {method}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Path
                  <input
                    value={smokeCheck.path}
                    onChange={(event) => updateSmokeCheck(smokeCheck.id, { path: event.target.value })}
                  />
                </label>
                <label>
                  Expected status
                  <input
                    max="599"
                    min="100"
                    type="number"
                    value={Number.isFinite(smokeCheck.expectedStatus) ? smokeCheck.expectedStatus : ""}
                    onChange={(event) =>
                      updateSmokeCheck(smokeCheck.id, {
                        expectedStatus: numberInputValue(event.currentTarget),
                      })
                    }
                  />
                </label>
              </div>
            </section>
          ))}
        </div>
      )}
    </form>
  );
}

function numberInputValue(input: HTMLInputElement): number {
  return input.value === "" ? Number.NaN : input.valueAsNumber;
}

function isHttpMethod(value: string): value is HttpMethod {
  return httpMethods.includes(value as HttpMethod);
}
