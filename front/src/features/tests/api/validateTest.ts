import type { ValidateResponse } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const validateTest = (testId: string) =>
  httpClient.post<ValidateResponse>(`/tests/${testId}/validate`);