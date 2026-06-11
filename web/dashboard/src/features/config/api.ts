import { request } from "../../shared/api/http";
import type { ConfigExplanation, ConfigSummary } from "./types";

export function getConfigSummary(): Promise<ConfigSummary> {
  return request<ConfigSummary>("/config/summary");
}

export function getConfigExplain(): Promise<ConfigExplanation> {
  return request<ConfigExplanation>("/config/explain");
}
