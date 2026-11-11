import type { CreateAttemptResponse } from "../../../entities/attempt";
import { httpClient } from "../../../shared/api/http";

export const createAttempt = (token: string, studentName: string) =>
  httpClient.post<CreateAttemptResponse>(`/public/tests/${token}/attempts`, {
    body: { studentName },
  });