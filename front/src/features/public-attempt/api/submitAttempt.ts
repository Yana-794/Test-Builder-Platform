import type { AttemptResultResponse } from "../../../entities/attempt";
import { httpClient } from "../../../shared/api/http";

export const submitAttempt = (attemptId: string) =>
  httpClient.post<AttemptResultResponse>(`/public/attempts/${attemptId}/submit`);