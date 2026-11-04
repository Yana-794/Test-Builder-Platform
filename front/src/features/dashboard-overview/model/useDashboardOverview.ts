import { useQuery } from "@tanstack/react-query";
import { fetchOverview } from "../api/fetchOverview";
import { ApiError } from "../../../shared/api/http";
import { dashboardQueryKeys } from "./queryKeys";

export function useDashboardOverview() {
  return useQuery({
    queryKey: dashboardQueryKeys.overview,
    queryFn: ({ signal }) => fetchOverview(signal),
    retry: (count, err) => {
      if (err instanceof ApiError) {
        if (err.status === 401 || err.status === 403) return false;
        if (err.status >= 400 && err.status < 500) return false;
      }
      return count < 2;
    },
    staleTime: 30_000,
  });
}