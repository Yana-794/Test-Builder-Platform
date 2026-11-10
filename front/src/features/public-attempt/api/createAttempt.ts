import type { CreateAttemptResponse } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const createAttempt = (token: string, studentName: string) =>
  httpClient.post<CreateAttemptResponse>(`/public/tests/${token}/attempts`, {
    body: { studentName },
  });