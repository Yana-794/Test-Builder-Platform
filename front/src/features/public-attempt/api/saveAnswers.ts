import type { SaveAnswersPayload } from "../../../entities/test/model/types";
import { httpClient } from "../../../shared/api/http";

export const saveAnswers = (attemptId: string, payload: SaveAnswersPayload) =>
  httpClient.put<void>(`/public/attempts/${attemptId}/answers`, { body: payload });