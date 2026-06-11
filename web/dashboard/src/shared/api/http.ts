type ApiErrorBody = {
  error?: string;
  message?: string;
};

const configuredBaseUrl = import.meta.env.VITE_PREDEPLOY_API_BASE_URL?.replace(/\/+$/, "") ?? "";

function buildApiUrl(path: string): string {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  if (!configuredBaseUrl) {
    return `/api${normalizedPath}`;
  }

  const baseIncludesApi = configuredBaseUrl.endsWith("/api");
  return `${configuredBaseUrl}${baseIncludesApi ? "" : "/api"}${normalizedPath}`;
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let response: Response;

  try {
    response = await fetch(buildApiUrl(path), {
      headers: {
        Accept: "application/json",
        ...init?.headers,
      },
      ...init,
    });
  } catch {
    throw new Error("Could not reach the PreDeploy Guard API. Start the backend and try again.");
  }

  if (!response.ok) {
    const message = await readErrorMessage(response);
    throw new Error(message || `API request failed with HTTP ${response.status}.`);
  }

  return response.json() as Promise<T>;
}

export async function requestText(path: string): Promise<string> {
  let response: Response;

  try {
    response = await fetch(buildApiUrl(path), {
      headers: {
        Accept: "text/markdown, text/plain, */*",
      },
    });
  } catch {
    throw new Error("Could not reach the PreDeploy Guard API. Start the backend and try again.");
  }

  if (!response.ok) {
    const message = await readTextErrorMessage(response);
    throw new Error(message || `API request failed with HTTP ${response.status}.`);
  }

  return response.text();
}

async function readErrorMessage(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as ApiErrorBody;
    return body.error ?? body.message ?? "";
  } catch {
    return "";
  }
}

async function readTextErrorMessage(response: Response): Promise<string> {
  const contentType = response.headers.get("Content-Type") ?? "";
  if (contentType.includes("application/json")) {
    return readErrorMessage(response);
  }

  try {
    return await response.text();
  } catch {
    return "";
  }
}
