import type { ChangeEvent } from "react";
import type { HttpMethod, PerformanceEndpointDraft } from "../types";

const httpMethods: HttpMethod[] = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"];

type PerformanceEndpointEditorProps = {
  endpoint: PerformanceEndpointDraft;
  onChange: (patch: Partial<PerformanceEndpointDraft>) => void;
  onRemove: () => void;
};

export function PerformanceEndpointEditor({ endpoint, onChange, onRemove }: PerformanceEndpointEditorProps) {
  function handleMethodChange(event: ChangeEvent<HTMLSelectElement>) {
    const method = event.target.value;
    if (isHttpMethod(method)) {
      onChange({ method });
    }
  }

  return (
    <section className="performance-endpoint-card">
      <div className="section-header">
        <h3>{endpoint.name || "Performance endpoint"}</h3>
        <button className="secondary-button" type="button" onClick={onRemove}>
          Remove
        </button>
      </div>

      <div className="form-grid">
        <label>
          Name
          <input value={endpoint.name} onChange={(event) => onChange({ name: event.target.value })} />
        </label>
        <label>
          Method
          <select value={endpoint.method} onChange={handleMethodChange}>
            {httpMethods.map((method) => (
              <option key={method} value={method}>
                {method}
              </option>
            ))}
          </select>
        </label>
        <label>
          Path
          <input value={endpoint.path} onChange={(event) => onChange({ path: event.target.value })} />
        </label>
      </div>
    </section>
  );
}

function isHttpMethod(value: string): value is HttpMethod {
  return httpMethods.includes(value as HttpMethod);
}
