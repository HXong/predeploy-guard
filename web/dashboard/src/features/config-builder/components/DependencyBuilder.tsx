import type { ChangeEvent, FormEvent } from "react";
import {
  createCustomDependency,
  createEnvVar,
  createPostgresDependency,
  createRedisDependency,
} from "../defaults";
import type { ConfigBuilderState, DependencyDraft, EnvVarDraft, ReadinessMode } from "../types";

const readinessModes: ReadinessMode[] = ["none", "shell", "command"];

type DependencyBuilderProps = {
  config: ConfigBuilderState;
  onChange: (config: ConfigBuilderState) => void;
};

export function DependencyBuilder({ config, onChange }: DependencyBuilderProps) {
  function updateDependencies(dependencies: DependencyDraft[]) {
    onChange({
      ...config,
      dependencies,
    });
  }

  function addDependency(dependency: DependencyDraft) {
    updateDependencies([...config.dependencies, dependency]);
  }

  function removeDependency(id: string) {
    updateDependencies(config.dependencies.filter((dependency) => dependency.id !== id));
  }

  function updateDependency(id: string, patch: Partial<DependencyDraft>) {
    updateDependencies(
      config.dependencies.map((dependency) => (dependency.id === id ? { ...dependency, ...patch } : dependency)),
    );
  }

  function updateReadiness(id: string, patch: Partial<DependencyDraft["readiness"]>) {
    updateDependencies(
      config.dependencies.map((dependency) =>
        dependency.id === id
          ? {
              ...dependency,
              readiness: {
                ...dependency.readiness,
                ...patch,
              },
            }
          : dependency,
      ),
    );
  }

  function updateEnv(dependency: DependencyDraft, envId: string, patch: Partial<EnvVarDraft>) {
    updateDependency(dependency.id, {
      env: dependency.env.map((item) => (item.id === envId ? { ...item, ...patch } : item)),
    });
  }

  function addEnv(dependency: DependencyDraft) {
    updateDependency(dependency.id, {
      env: [...dependency.env, createEnvVar()],
    });
  }

  function removeEnv(dependency: DependencyDraft, envId: string) {
    updateDependency(dependency.id, {
      env: dependency.env.filter((item) => item.id !== envId),
    });
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <form className="builder-form" onSubmit={handleSubmit}>
      <div className="profile-buttons">
        <button className="secondary-button" type="button" onClick={() => addDependency(createCustomDependency())}>
          Add custom dependency
        </button>
        <button className="secondary-button" type="button" onClick={() => addDependency(createPostgresDependency())}>
          Add PostgreSQL
        </button>
        <button className="secondary-button" type="button" onClick={() => addDependency(createRedisDependency())}>
          Add Redis
        </button>
      </div>

      {config.dependencies.length === 0 ? (
        <p className="empty-state">No dependencies configured.</p>
      ) : (
        <div className="dependency-list">
          {config.dependencies.map((dependency) => (
            <section className="dependency-card" key={dependency.id}>
              <div className="section-header">
                <h3>{dependency.name || "Dependency"}</h3>
                <button className="secondary-button" type="button" onClick={() => removeDependency(dependency.id)}>
                  Remove
                </button>
              </div>

              <div className="form-grid">
                <label>
                  Name
                  <input
                    value={dependency.name}
                    onChange={(event) => updateDependency(dependency.id, { name: event.target.value })}
                  />
                </label>
                <label>
                  Image
                  <input
                    value={dependency.image}
                    onChange={(event) => updateDependency(dependency.id, { image: event.target.value })}
                  />
                </label>
                <label>
                  Port
                  <input
                    min="1"
                    type="number"
                    value={Number.isFinite(dependency.port) ? dependency.port : ""}
                    onChange={(event) =>
                      updateDependency(dependency.id, { port: optionalNumberInputValue(event.currentTarget) })
                    }
                  />
                </label>
                <label>
                  Readiness
                  <select
                    value={dependency.readiness.mode}
                    onChange={(event) =>
                      updateReadinessMode(event, (mode) => updateReadiness(dependency.id, { mode }))
                    }
                  >
                    {readinessModes.map((mode) => (
                      <option key={mode} value={mode}>
                        {mode}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              {dependency.readiness.mode !== "none" && (
                <div className="form-grid">
                  {dependency.readiness.mode === "shell" && (
                    <label>
                      Shell command
                      <input
                        value={dependency.readiness.shell}
                        onChange={(event) => updateReadiness(dependency.id, { shell: event.target.value })}
                      />
                    </label>
                  )}
                  {dependency.readiness.mode === "command" && (
                    <label>
                      Command
                      <input
                        value={dependency.readiness.command}
                        onChange={(event) => updateReadiness(dependency.id, { command: event.target.value })}
                      />
                    </label>
                  )}
                  <label>
                    Interval seconds
                    <input
                      min="1"
                      type="number"
                      value={
                        Number.isFinite(dependency.readiness.intervalSeconds)
                          ? dependency.readiness.intervalSeconds
                          : ""
                      }
                      onChange={(event) =>
                        updateReadiness(dependency.id, { intervalSeconds: requiredNumberInputValue(event.currentTarget) })
                      }
                    />
                  </label>
                  <label>
                    Timeout seconds
                    <input
                      min="1"
                      type="number"
                      value={
                        Number.isFinite(dependency.readiness.timeoutSeconds)
                          ? dependency.readiness.timeoutSeconds
                          : ""
                      }
                      onChange={(event) =>
                        updateReadiness(dependency.id, { timeoutSeconds: requiredNumberInputValue(event.currentTarget) })
                      }
                    />
                  </label>
                </div>
              )}

              <div className="env-editor">
                <div className="section-header">
                  <h3>Environment</h3>
                  <button className="secondary-button" type="button" onClick={() => addEnv(dependency)}>
                    Add env var
                  </button>
                </div>
                {dependency.env.length === 0 ? (
                  <p className="empty-state">No environment variables.</p>
                ) : (
                  <div className="env-rows">
                    {dependency.env.map((item) => (
                      <div className="env-row" key={item.id}>
                        <label>
                          Key
                          <input
                            value={item.key}
                            onChange={(event) => updateEnv(dependency, item.id, { key: event.target.value })}
                          />
                        </label>
                        <label>
                          Value
                          <input
                            value={item.value}
                            onChange={(event) => updateEnv(dependency, item.id, { value: event.target.value })}
                          />
                        </label>
                        <button className="secondary-button" type="button" onClick={() => removeEnv(dependency, item.id)}>
                          Remove
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </section>
          ))}
        </div>
      )}
    </form>
  );
}

function updateReadinessMode(
  event: ChangeEvent<HTMLSelectElement>,
  onChange: (mode: ReadinessMode) => void,
) {
  const value = event.target.value;
  if (isReadinessMode(value)) {
    onChange(value);
  }
}

function optionalNumberInputValue(input: HTMLInputElement): number | undefined {
  return input.value === "" ? undefined : input.valueAsNumber;
}

function requiredNumberInputValue(input: HTMLInputElement): number {
  return input.value === "" ? Number.NaN : input.valueAsNumber;
}

function isReadinessMode(value: string): value is ReadinessMode {
  return readinessModes.includes(value as ReadinessMode);
}
