import type { CreateTestPayload, Test } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

interface CreateTestResponse {
  test: Test;
}

export const createTest = (payload: CreateTestPayload) =>
  httpClient.post<CreateTestResponse>("/tests",  { body: payload });