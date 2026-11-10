import type { AttemptResultResponse } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const fetchAttemptResult = (attemptId: string, signal?: AbortSignal) =>
  httpClient.get<AttemptResultResponse>(`/public/attempts/${attemptId}/result`, { signal });