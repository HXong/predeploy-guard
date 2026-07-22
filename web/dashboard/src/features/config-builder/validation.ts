import type { ConfigBuilderState, HttpMethod, PerformanceConfigDraft, ValidationError } from "./types";

const supportedHttpMethods: HttpMethod[] = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"];
const positiveGoDurationPattern = /^(?:(?:\d+(?:\.\d+)?|\.\d+)(?:ns|us|\u00b5s|\u03bcs|ms|s|m|h))+$/;
const profileNamePattern = /^[A-Za-z0-9][A-Za-z0-9_-]*$/;

export function validateConfigBuilder(config: ConfigBuilderState): ValidationError[] {
  const errors: ValidationError[] = [];

  validateService(config, errors);
  validateDependencies(config, errors);
  validateSmokeChecks(config, errors);
  validatePerformanceConfig(config.performance, config.checks.smoke.length, "performance", errors);
  validateProfiles(config, errors);

  if (!Number.isFinite(config.settings.timeoutSeconds) || config.settings.timeoutSeconds <= 0) {
    errors.push({ field: "settings.timeoutSeconds", message: "Timeout must be a positive number of seconds." });
  }

  return errors;
}

function validateService(config: ConfigBuilderState, errors: ValidationError[]) {
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

function validateProfiles(config: ConfigBuilderState, errors: ValidationError[]) {
  const seenProfileNames = new Set<string>();

  config.profiles.forEach((profile, profileIndex) => {
    const name = profile.name.trim();
    const nameField = `profiles.${profileIndex}.name`;

    if (name === "") {
      errors.push({ field: nameField, message: "Profile name is required." });
    } else {
      if (seenProfileNames.has(name)) {
        errors.push({ field: nameField, message: `Profile ${name} is duplicated.` });
      }

      if (name.toLowerCase() === "base") {
        errors.push({ field: nameField, message: "Profile name cannot be base." });
      }

      if (!profileNamePattern.test(name)) {
        errors.push({
          field: nameField,
          message: "Profile name must start with a letter or number and contain only letters, numbers, underscores, or hyphens.",
        });
      }

      seenProfileNames.add(name);
    }

    validatePerformanceConfig(
      profile.performance,
      config.checks.smoke.length,
      `profiles.${profileIndex}.performance`,
      errors,
    );
  });
}

function validatePerformanceConfig(
  performance: PerformanceConfigDraft,
  smokeCheckCount: number,
  fieldPrefix: string,
  errors: ValidationError[],
) {
  if (!performance.enabled) {
    return;
  }

  if (!Number.isInteger(performance.vus) || performance.vus <= 0) {
    errors.push({ field: `${fieldPrefix}.vus`, message: "Performance VUs must be a positive integer." });
  }

  if (performance.duration.trim() === "") {
    errors.push({ field: `${fieldPrefix}.duration`, message: "Performance duration is required." });
  } else if (!isPositiveGoDuration(performance.duration.trim())) {
    errors.push({
      field: `${fieldPrefix}.duration`,
      message: "Performance duration must be a positive duration such as 250ms, 15s, 1m, or 1m30s.",
    });
  }

  if (
    !Number.isFinite(performance.thresholds.maxP95LatencyMs) ||
    performance.thresholds.maxP95LatencyMs <= 0
  ) {
    errors.push({
      field: `${fieldPrefix}.thresholds.maxP95LatencyMs`,
      message: "Maximum p95 latency must be greater than 0 milliseconds.",
    });
  }

  if (
    !Number.isFinite(performance.thresholds.maxErrorRate) ||
    performance.thresholds.maxErrorRate <= 0 ||
    performance.thresholds.maxErrorRate > 1
  ) {
    errors.push({
      field: `${fieldPrefix}.thresholds.maxErrorRate`,
      message: "Maximum error rate must be greater than 0 and at most 1.",
    });
  }

  if (performance.endpoints.length === 0) {
    if (smokeCheckCount === 0) {
      errors.push({
        field: `${fieldPrefix}.endpoints`,
        message: "Add at least one performance endpoint or smoke check while performance testing is enabled.",
      });
    }
    return;
  }

  const seenEndpointNames = new Set<string>();
  performance.endpoints.forEach((endpoint, endpointIndex) => {
    const name = endpoint.name.trim();
    const label = name || String(endpointIndex + 1);

    if (name === "") {
      errors.push({
        field: `${fieldPrefix}.endpoints.${endpointIndex}.name`,
        message: "Performance endpoint name is required.",
      });
    } else if (seenEndpointNames.has(name)) {
      errors.push({
        field: `${fieldPrefix}.endpoints.${endpointIndex}.name`,
        message: `Performance endpoint ${name} is duplicated.`,
      });
    }

    if (name !== "") {
      seenEndpointNames.add(name);
    }

    if (!isSupportedHttpMethod(endpoint.method)) {
      errors.push({
        field: `${fieldPrefix}.endpoints.${endpointIndex}.method`,
        message: `Performance endpoint ${label} uses an unsupported HTTP method.`,
      });
    }

    if (endpoint.path.trim() === "" || !endpoint.path.trim().startsWith("/")) {
      errors.push({
        field: `${fieldPrefix}.endpoints.${endpointIndex}.path`,
        message: `Performance endpoint ${label} path should start with /.`,
      });
    }
  });
}

function isPositiveGoDuration(value: string): boolean {
  if (!positiveGoDurationPattern.test(value)) {
    return false;
  }

  const numericParts = value.match(/\d+(?:\.\d+)?|\.\d+/g) ?? [];
  return numericParts.some((part) => Number(part) > 0);
}

function isSupportedHttpMethod(value: string): value is HttpMethod {
  return supportedHttpMethods.includes(value as HttpMethod);
}
