import type { ConfigBuilderState, EnvVarDraft } from "./types";

export function generatePredeployYaml(config: ConfigBuilderState): string {
  const lines = [
    "runtime:",
    `  type: ${yamlScalar(config.runtime.type)}`,
    "",
    "service:",
    `  name: ${yamlScalar(config.service.name)}`,
    `  image: ${yamlScalar(config.service.image)}`,
    "  build:",
    `    context: ${yamlScalar(config.service.build.context)}`,
    `    dockerfile: ${yamlScalar(config.service.build.dockerfile)}`,
    `  port: ${normalizeNumber(config.service.port)}`,
    `  healthPath: ${yamlScalar(config.service.healthPath)}`,
  ];

  const env = config.service.env.filter((item) => item.key.trim() !== "");
  if (env.length > 0) {
    lines.push("  env:");
    for (const item of env) {
      lines.push(`    ${yamlKey(item.key)}: ${yamlQuotedString(item.value)}`);
    }
  }

  lines.push(
    "",
    "settings:",
    `  cleanup: ${config.settings.cleanup ? "true" : "false"}`,
    `  timeoutSeconds: ${normalizeNumber(config.settings.timeoutSeconds)}`,
    "",
  );

  return lines.join("\n");
}

function normalizeNumber(value: number): string {
  if (!Number.isFinite(value)) {
    return "0";
  }

  return String(value);
}

function yamlKey(value: string): string {
  const trimmed = value.trim();
  if (/^[A-Za-z_][A-Za-z0-9_]*$/.test(trimmed)) {
    return trimmed;
  }

  return yamlQuotedString(trimmed);
}

function yamlScalar(value: string): string {
  const trimmed = value.trim();
  if (trimmed === "") {
    return yamlQuotedString("");
  }

  if (isPlainSafeScalar(trimmed)) {
    return trimmed;
  }

  return yamlQuotedString(trimmed);
}

function yamlQuotedString(value: string): string {
  return JSON.stringify(value);
}

function isPlainSafeScalar(value: string): boolean {
  if (!/^[A-Za-z0-9_./:-]+$/.test(value)) {
    return false;
  }

  const lower = value.toLowerCase();
  return !["true", "false", "null", "~"].includes(lower) && Number.isNaN(Number(value));
}

export function envRowsToRecord(env: EnvVarDraft[]): Record<string, string> {
  return env.reduce<Record<string, string>>((result, item) => {
    const key = item.key.trim();
    if (key !== "") {
      result[key] = item.value;
    }

    return result;
  }, {});
}
