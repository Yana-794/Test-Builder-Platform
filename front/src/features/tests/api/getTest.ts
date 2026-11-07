import type { TestEditor } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

interface GetTestResponse {
  test: TestEditor;
}

export const getTest = (testId: string, signal?: AbortSignal) =>
  httpClient.get<GetTestResponse>(`/tests/${testId}`, { signal });