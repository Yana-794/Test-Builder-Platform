import type { AnalyticsRecentAttempt } from "../../../entities/analytics";
import { httpClient } from "../../../shared/api/http";

type Response = { attempts: AnalyticsRecentAttempt[] };

export const fetchRecentAttempts = (limit: number, signal?: AbortSignal) =>
  httpClient.get<Response>(`/analytics/recent?limit=${limit}`, { signal });