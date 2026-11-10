import type { TestFilter, TestListResponse } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const fetchTests = (filter: TestFilter, signal?: AbortSignal) =>
  httpClient.get<TestListResponse>(`/tests?filter=${filter}`, { signal });