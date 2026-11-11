import { useQuery } from "@tanstack/react-query";
import { fetchRecentAttempts } from "../api/fetchRecentAttempts";
import { analyticsQueryKeys } from "./queryKeys";
import { ApiError } from "../../../shared/api/http";

export function useRecentAttempts(limit = 50) {
  return useQuery({
    queryKey: analyticsQueryKeys.recent(limit),
    queryFn: ({ signal }) => fetchRecentAttempts(limit, signal),
    staleTime: 30_000,
    retry: (count, err) => {
      if (err instanceof ApiError && err.status === 401) return false;
      return count < 2;
    },
  });
}