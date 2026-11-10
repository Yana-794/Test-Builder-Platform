import { useQuery } from "@tanstack/react-query";
import { fetchAttemptState } from "../api/fetchAttemptState";
import { attemptQueryKeys } from "./queryKeys";

export function useAttemptState(attemptId: string) {
  return useQuery({
    queryKey: attemptQueryKeys.state(attemptId),
    queryFn: ({ signal }) => fetchAttemptState(attemptId, signal),
    staleTime: Infinity,
    retry: false,
    enabled: !!attemptId,
  });
}