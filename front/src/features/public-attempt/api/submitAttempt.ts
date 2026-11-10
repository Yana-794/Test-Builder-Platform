import type { AttemptResultResponse } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const submitAttempt = (attemptId: string) =>
  httpClient.post<AttemptResultResponse>(`/public/attempts/${attemptId}/submit`);