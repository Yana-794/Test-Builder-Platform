import type { ValidateResponse } from "../../../entities/publication";
import { httpClient } from "../../../shared/api/http";

export const validateTest = (testId: string) =>
  httpClient.post<ValidateResponse>(`/tests/${testId}/validate`);