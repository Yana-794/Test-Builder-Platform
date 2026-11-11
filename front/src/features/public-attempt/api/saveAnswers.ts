import type { SaveAnswersPayload } from "../../../entities/attempt";
import { httpClient } from "../../../shared/api/http";

export const saveAnswers = (attemptId: string, payload: SaveAnswersPayload) =>
  httpClient.put<void>(`/public/attempts/${attemptId}/answers`, { body: payload });