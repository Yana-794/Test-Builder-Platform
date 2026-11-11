import { useMutation } from "@tanstack/react-query";
import { saveAnswers } from "../api/saveAnswers";
import type { SaveAnswersPayload } from "../../../entities/attempt";

export function useSaveAnswers(attemptId: string) {
  return useMutation({
    mutationFn: (payload: SaveAnswersPayload) => saveAnswers(attemptId, payload),
    retry: 2,
  });
}