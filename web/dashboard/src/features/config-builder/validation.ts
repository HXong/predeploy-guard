import type { ConfigBuilderState, ValidationError } from "./types";

export function validateConfigBuilder(config: ConfigBuilderState): ValidationError[] {
  const errors: ValidationError[] = [];

  if (config.service.name.trim() === "") {
    errors.push({ field: "service.name", message: "Service name is required." });
  }

  if (config.service.image.trim() === "") {
    errors.push({ field: "service.image", message: "Service image is required." });
  }

  if (!Number.isFinite(config.service.port) || config.service.port <= 0) {
    errors.push({ field: "service.port", message: "Service port must be a positive number." });
  }

  if (config.service.healthPath.trim() === "" || !config.service.healthPath.trim().startsWith("/")) {
    errors.push({ field: "service.healthPath", message: "Health path should start with /." });
  }

  if (!Number.isFinite(config.settings.timeoutSeconds) || config.settings.timeoutSeconds <= 0) {
    errors.push({ field: "settings.timeoutSeconds", message: "Timeout must be a positive number of seconds." });
  }

  const seenEnvKeys = new Set<string>();
  config.service.env.forEach((item, index) => {
    const key = item.key.trim();
    if (key === "") {
      errors.push({ field: `service.env.${index}.key`, message: "Environment variable keys cannot be empty." });
      return;
    }

    if (seenEnvKeys.has(key)) {
      errors.push({ field: `service.env.${index}.key`, message: `Environment variable ${key} is duplicated.` });
    }

    seenEnvKeys.add(key);
  });

  validateDependencies(config, errors);
  validateSmokeChecks(config, errors);

  return errors;
}

function validateSmokeChecks(config: ConfigBuilderState, errors: ValidationError[]) {
  const seenSmokeCheckNames = new Set<string>();

  config.checks.smoke.forEach((smokeCheck, smokeCheckIndex) => {
    const name = smokeCheck.name.trim();
    const label = name || String(smokeCheckIndex + 1);

    if (name === "") {
      errors.push({
        field: `checks.smoke.${smokeCheckIndex}.name`,
        message: "Smoke check name is required.",
      });
    } else if (seenSmokeCheckNames.has(name)) {
      errors.push({
        field: `checks.smoke.${smokeCheckIndex}.name`,
        message: `Smoke check ${name} is duplicated.`,
      });
    }

    if (name !== "") {
      seenSmokeCheckNames.add(name);
    }

    if (smokeCheck.path.trim() === "" || !smokeCheck.path.trim().startsWith("/")) {
      errors.push({
        field: `checks.smoke.${smokeCheckIndex}.path`,
        message: `Smoke check ${label} path should start with /.`,
      });
    }

    if (
      !Number.isInteger(smokeCheck.expectedStatus) ||
      smokeCheck.expectedStatus < 100 ||
      smokeCheck.expectedStatus > 599
    ) {
      errors.push({
        field: `checks.smoke.${smokeCheckIndex}.expectedStatus`,
        message: `Smoke check ${label} expected status must be between 100 and 599.`,
      });
    }
  });
}

function validateDependencies(config: ConfigBuilderState, errors: ValidationError[]) {
  const seenDependencyNames = new Set<string>();

  config.dependencies.forEach((dependency, dependencyIndex) => {
    const name = dependency.name.trim();
    if (name === "") {
      errors.push({
        field: `dependencies.${dependencyIndex}.name`,
        message: "Dependency name is required.",
      });
    } else if (seenDependencyNames.has(name)) {
      errors.push({
        field: `dependencies.${dependencyIndex}.name`,
        message: `Dependency ${name} is duplicated.`,
      });
    }

    if (name !== "") {
      seenDependencyNames.add(name);
    }

    if (dependency.image.trim() === "") {
      errors.push({
        field: `dependencies.${dependencyIndex}.image`,
        message: `Dependency ${name || dependencyIndex + 1} image is required.`,
      });
    }

    if (dependency.port !== undefined && (!Number.isFinite(dependency.port) || dependency.port <= 0)) {
      errors.push({
        field: `dependencies.${dependencyIndex}.port`,
        message: `Dependency ${name || dependencyIndex + 1} port must be positive when provided.`,
      });
    }

    validateDependencyEnv(dependency.env, dependencyIndex, name || String(dependencyIndex + 1), errors);
    validateReadiness(config, dependencyIndex, name || String(dependencyIndex + 1), errors);
  });
}

function validateDependencyEnv(
  env: ConfigBuilderState["dependencies"][number]["env"],
  dependencyIndex: number,
  dependencyLabel: string,
  errors: ValidationError[],
) {
  const seenEnvKeys = new Set<string>();

  env.forEach((item, envIndex) => {
    const key = item.key.trim();
    if (key === "") {
      errors.push({
        field: `dependencies.${dependencyIndex}.env.${envIndex}.key`,
        message: `Dependency ${dependencyLabel} environment variable keys cannot be empty.`,
      });
      return;
    }

    if (seenEnvKeys.has(key)) {
      errors.push({
        field: `dependencies.${dependencyIndex}.env.${envIndex}.key`,
        message: `Dependency ${dependencyLabel} environment variable ${key} is duplicated.`,
      });
    }

    seenEnvKeys.add(key);
  });
}

function validateReadiness(
  config: ConfigBuilderState,
  dependencyIndex: number,
  dependencyLabel: string,
  errors: ValidationError[],
) {
  const readiness = config.dependencies[dependencyIndex].readiness;
  if (readiness.mode === "none") {
    return;
  }

  if (!Number.isFinite(readiness.intervalSeconds) || readiness.intervalSeconds <= 0) {
    errors.push({
      field: `dependencies.${dependencyIndex}.readiness.intervalSeconds`,
      message: `Dependency ${dependencyLabel} readiness interval must be positive.`,
    });
  }

  if (!Number.isFinite(readiness.timeoutSeconds) || readiness.timeoutSeconds <= 0) {
    errors.push({
      field: `dependencies.${dependencyIndex}.readiness.timeoutSeconds`,
      message: `Dependency ${dependencyLabel} readiness timeout must be positive.`,
    });
  }

  if (readiness.mode === "shell" && readiness.shell.trim() === "") {
    errors.push({
      field: `dependencies.${dependencyIndex}.readiness.shell`,
      message: `Dependency ${dependencyLabel} shell readiness command is required.`,
    });
  }

  if (readiness.mode === "command" && readiness.command.trim() === "") {
    errors.push({
      field: `dependencies.${dependencyIndex}.readiness.command`,
      message: `Dependency ${dependencyLabel} readiness command is required.`,
    });
  }
}
