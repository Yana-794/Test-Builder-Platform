import { useQuery } from "@tanstack/react-query";
import { fetchPublicTest } from "../api/fetchPublicTest";
import { attemptQueryKeys } from "./queryKeys";

export function usePublicTest(token: string) {
  return useQuery({
    queryKey: attemptQueryKeys.publicTest(token),
    queryFn: ({ signal }) => fetchPublicTest(token, signal),
    staleTime: 60_000,
    retry: false,
    enabled: !!token,
  });
}