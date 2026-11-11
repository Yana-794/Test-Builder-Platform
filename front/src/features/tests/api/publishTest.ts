import type { PublicationResponse } from "../../../entities/publication";
import { httpClient } from "../../../shared/api/http";

export const publishTest = (testId: string, expectedRevision: number) =>
  httpClient.post<PublicationResponse>(`/tests/${testId}/publish`, {
    body: { expectedRevision },
  });
