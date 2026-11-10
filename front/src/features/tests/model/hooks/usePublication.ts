import { useQuery } from "@tanstack/react-query";
import { ApiError } from "../../../../shared/api/http";
import { fetchPublication } from "../../api/fetchPublication";


export function usePublication(testId: string, enabled: boolean) {
  return useQuery({
    queryKey: ["tests", "publication", testId] as const,
    queryFn: ({ signal }) => fetchPublication(testId, signal),
    enabled: enabled && !!testId,
    staleTime: 60_000,
    retry: (count, err) => {
      if (err instanceof ApiError && err.status === 404) return false;
      return count < 2;
    },
  });
}