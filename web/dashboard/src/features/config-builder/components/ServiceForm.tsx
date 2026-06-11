import type { ChangeEvent, FormEvent } from "react";
import type { ConfigBuilderState, EnvVarDraft, RuntimeType } from "../types";

const runtimeTypes: RuntimeType[] = ["docker-compose"];

type ServiceFormProps = {
  config: ConfigBuilderState;
  onChange: (config: ConfigBuilderState) => void;
};

export function ServiceForm({ config, onChange }: ServiceFormProps) {
  function updateService<Field extends keyof ConfigBuilderState["service"]>(
    field: Field,
    value: ConfigBuilderState["service"][Field],
  ) {
    onChange({
      ...config,
      service: {
        ...config.service,
        [field]: value,
      },
    });
  }

  function updateBuild<Field extends keyof ConfigBuilderState["service"]["build"]>(
    field: Field,
    value: ConfigBuilderState["service"]["build"][Field],
  ) {
    onChange({
      ...config,
      service: {
        ...config.service,
        build: {
          ...config.service.build,
          [field]: value,
        },
      },
    });
  }

  function updateRuntimeType(type: RuntimeType) {
    onChange({
      ...config,
      runtime: {
        ...config.runtime,
        type,
      },
    });
  }

  function updateSettings(settings: Partial<ConfigBuilderState["settings"]>) {
    onChange({
      ...config,
      settings: {
        ...config.settings,
        ...settings,
      },
    });
  }

  function updateEnv(id: string, patch: Partial<EnvVarDraft>) {
    updateService(
      "env",
      config.service.env.map((item) => (item.id === id ? { ...item, ...patch } : item)),
    );
  }

  function addEnv() {
    updateService("env", [
      ...config.service.env,
      {
        id: `env-${Date.now()}`,
        key: "",
        value: "",
      },
    ]);
  }

  function removeEnv(id: string) {
    updateService(
      "env",
      config.service.env.filter((item) => item.id !== id),
    );
  }

  function handleRuntimeChange(event: ChangeEvent<HTMLSelectElement>) {
    const value = event.target.value;
    if (isRuntimeType(value)) {
      updateRuntimeType(value);
    }
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <form className="builder-form" onSubmit={handleSubmit}>
      <div className="form-grid">
        <label>
          Runtime
          <select value={config.runtime.type} onChange={handleRuntimeChange}>
            {runtimeTypes.map((runtimeType) => (
              <option key={runtimeType} value={runtimeType}>
                {runtimeType}
              </option>
            ))}
          </select>
        </label>
        <label>
          Service name
          <input value={config.service.name} onChange={(event) => updateService("name", event.target.value)} />
        </label>
        <label>
          Image
          <input value={config.service.image} onChange={(event) => updateService("image", event.target.value)} />
        </label>
        <label>
          Port
          <input
            min="1"
            type="number"
            value={Number.isFinite(config.service.port) ? config.service.port : ""}
            onChange={(event) => updateService("port", numberInputValue(event.currentTarget))}
          />
        </label>
        <label>
          Health path
          <input
            value={config.service.healthPath}
            onChange={(event) => updateService("healthPath", event.target.value)}
          />
        </label>
        <label>
          Build context
          <input
            value={config.service.build.context}
            onChange={(event) => updateBuild("context", event.target.value)}
          />
        </label>
        <label>
          Dockerfile
          <input
            value={config.service.build.dockerfile}
            onChange={(event) => updateBuild("dockerfile", event.target.value)}
          />
        </label>
        <label>
          Timeout seconds
          <input
            min="1"
            type="number"
            value={Number.isFinite(config.settings.timeoutSeconds) ? config.settings.timeoutSeconds : ""}
            onChange={(event) => updateSettings({ timeoutSeconds: numberInputValue(event.currentTarget) })}
          />
        </label>
      </div>

      <label className="checkbox-row">
        <input
          checked={config.settings.cleanup}
          type="checkbox"
          onChange={(event) => updateSettings({ cleanup: event.target.checked })}
        />
        Clean up sandbox after each run
      </label>

      <div className="env-editor">
        <div className="section-header">
          <h3>Environment</h3>
          <button className="secondary-button" type="button" onClick={addEnv}>
            Add env var
          </button>
        </div>
        <div className="env-rows">
          {config.service.env.map((item) => (
            <div className="env-row" key={item.id}>
              <label>
                Key
                <input value={item.key} onChange={(event) => updateEnv(item.id, { key: event.target.value })} />
              </label>
              <label>
                Value
                <input value={item.value} onChange={(event) => updateEnv(item.id, { value: event.target.value })} />
              </label>
              <button className="secondary-button" type="button" onClick={() => removeEnv(item.id)}>
                Remove
              </button>
            </div>
          ))}
        </div>
      </div>
    </form>
  );
}

function numberInputValue(input: HTMLInputElement): number {
  return input.value === "" ? Number.NaN : input.valueAsNumber;
}

function isRuntimeType(value: string): value is RuntimeType {
  return runtimeTypes.includes(value as RuntimeType);
}
