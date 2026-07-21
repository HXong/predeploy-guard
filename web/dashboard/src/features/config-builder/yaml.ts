import type { ConfigBuilderState, DependencyDraft, EnvVarDraft, SmokeCheckDraft } from "./types";

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

  if (config.dependencies.length > 0) {
    lines.push("", "dependencies:");
    config.dependencies.forEach((dependency, index) => {
      appendDependencyYaml(lines, dependency, index);
    });
  }

  if (config.checks.smoke.length > 0) {
    lines.push("", "checks:", "  smoke:");
    for (const smokeCheck of config.checks.smoke) {
      appendSmokeCheckYaml(lines, smokeCheck);
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

function appendSmokeCheckYaml(lines: string[], smokeCheck: SmokeCheckDraft) {
  lines.push(`    - name: ${yamlScalar(smokeCheck.name)}`);
  lines.push(`      method: ${smokeCheck.method}`);
  lines.push(`      path: ${yamlScalar(smokeCheck.path)}`);
  lines.push(`      expectedStatus: ${normalizeNumber(smokeCheck.expectedStatus)}`);
}

function appendDependencyYaml(lines: string[], dependency: DependencyDraft, index: number) {
  if (index > 0) {
    lines.push("");
  }

  lines.push(`  ${yamlKey(dependency.name)}:`);
  lines.push(`    image: ${yamlScalar(dependency.image)}`);

  if (dependency.port !== undefined && Number.isFinite(dependency.port) && dependency.port > 0) {
    lines.push(`    port: ${normalizeNumber(dependency.port)}`);
  }

  const env = dependency.env.filter((item) => item.key.trim() !== "");
  if (env.length > 0) {
    lines.push("    env:");
    for (const item of env) {
      lines.push(`      ${yamlKey(item.key)}: ${yamlQuotedString(item.value)}`);
    }
  }

  if (dependency.readiness.mode === "shell" && dependency.readiness.shell.trim() !== "") {
    lines.push("    readiness:");
    lines.push(`      shell: ${yamlQuotedString(dependency.readiness.shell)}`);
    lines.push(`      intervalSeconds: ${normalizeNumber(dependency.readiness.intervalSeconds)}`);
    lines.push(`      timeoutSeconds: ${normalizeNumber(dependency.readiness.timeoutSeconds)}`);
  }

  if (dependency.readiness.mode === "command" && dependency.readiness.command.trim() !== "") {
    lines.push("    readiness:");
    lines.push(`      command: ${yamlCommandArray(dependency.readiness.command)}`);
    lines.push(`      intervalSeconds: ${normalizeNumber(dependency.readiness.intervalSeconds)}`);
    lines.push(`      timeoutSeconds: ${normalizeNumber(dependency.readiness.timeoutSeconds)}`);
  }
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

function yamlCommandArray(value: string): string {
  const parts = value.trim().split(/\s+/).filter(Boolean);
  return `[${parts.map(yamlQuotedString).join(", ")}]`;
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
