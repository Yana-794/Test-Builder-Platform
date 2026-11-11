import type { PublicationResponse } from "../../../entities/publication";
import { httpClient } from "../../../shared/api/http";

export const fetchPublication = (testId: string, signal?: AbortSignal) =>
  httpClient.get<PublicationResponse>(`/tests/${testId}/publication`, { signal });