import type { PublicationResponse } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const fetchPublication = (testId: string, signal?: AbortSignal) =>
  httpClient.get<PublicationResponse>(`/tests/${testId}/publication`, { signal });