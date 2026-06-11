import type { ValidationError } from "../types";

type YamlPreviewProps = {
  errors: ValidationError[];
  yaml: string;
};

export function YamlPreview({ errors, yaml }: YamlPreviewProps) {
  async function copyYaml() {
    if (navigator.clipboard) {
      await navigator.clipboard.writeText(yaml);
      return;
    }

    const textarea = document.createElement("textarea");
    textarea.value = yaml;
    textarea.style.position = "fixed";
    textarea.style.left = "-9999px";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
  }

  function downloadYaml() {
    const blob = new Blob([yaml], { type: "text/yaml;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = "predeploy.yaml";
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
  }

  return (
    <div className="yaml-preview-panel">
      <div className="panel-actions">
        <button className="secondary-button" type="button" onClick={copyYaml}>
          Copy YAML
        </button>
        <button className="primary-button" type="button" onClick={downloadYaml}>
          Download YAML
        </button>
      </div>

      {errors.length > 0 && (
        <div className="validation-panel" role="alert">
          <strong>Fix these before using the config:</strong>
          <ul>
            {errors.map((error) => (
              <li key={`${error.field}-${error.message}`}>{error.message}</li>
            ))}
          </ul>
        </div>
      )}

      <pre className="yaml-preview">{yaml}</pre>
    </div>
  );
}
