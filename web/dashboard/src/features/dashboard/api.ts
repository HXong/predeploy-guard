import { request } from "../../shared/api/http";
import type { HealthResponse } from "./types";

export function getHealth(): Promise<HealthResponse> {
  return request<HealthResponse>("/health");
}
