
import type { AttemptResultResponse } from "../../../entities/attempt";
import { httpClient } from "../../../shared/api/http";

export const fetchAttemptResult = (attemptId: string, signal?: AbortSignal) =>
  httpClient.get<AttemptResultResponse>(`/public/attempts/${attemptId}/result`, { signal });