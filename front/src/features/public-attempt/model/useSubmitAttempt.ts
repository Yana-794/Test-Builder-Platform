import { useMutation, useQueryClient } from "@tanstack/react-query";
import { submitAttempt } from "../api/submitAttempt";
import { attemptQueryKeys } from "./queryKeys";

export function useSubmitAttempt(attemptId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => submitAttempt(attemptId),
    retry: 1,
    onSuccess: (data) => {
      queryClient.setQueryData(attemptQueryKeys.result(attemptId), data);
    },
  });
}