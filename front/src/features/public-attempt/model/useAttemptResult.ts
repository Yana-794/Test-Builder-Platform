import { useQuery } from "@tanstack/react-query";
import { fetchAttemptResult } from "../api/fetchAttemptResult";
import { attemptQueryKeys } from "./queryKeys";

export function useAttemptResult(attemptId: string) {
  return useQuery({
    queryKey: attemptQueryKeys.result(attemptId),
    queryFn: ({ signal }) => fetchAttemptResult(attemptId, signal),
    staleTime: Infinity,
    retry: false,
    enabled: !!attemptId,
  });
}