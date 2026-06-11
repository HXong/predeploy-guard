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

  return errors;
}
