import type { DashboardOverview } from "../../../entities/dashboard";
import { httpClient } from "../../../shared/api/http";

export const fetchOverview = (signal?: AbortSignal) =>
  httpClient.get<DashboardOverview>("/app/overview", { signal });