import type { SaveDraftPayload, TestEditor } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

interface SaveDraftResponse {
  test: TestEditor;
}

export const saveDraft = (testId: string, payload: SaveDraftPayload) =>
  httpClient.put<SaveDraftResponse>(`/tests/${testId}/draft`,   { body: payload });