
import type { AttemptStateResponse } from "../../../entities/attempt";
import { httpClient } from "../../../shared/api/http";

export const fetchAttemptState = (attemptId: string, signal?: AbortSignal) =>
  httpClient.get<AttemptStateResponse>(`/public/attempts/${attemptId}`, { signal });