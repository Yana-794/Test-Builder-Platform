import { useQuery } from "@tanstack/react-query";
import { testsQueryKeys } from "../queryKeys";
import type { TestFilter } from "../../../../entities/test/model/types";
import { fetchTests } from "../../api/fetchTests";
;

export function useTests(filter: TestFilter) {
  return useQuery({
    queryKey: testsQueryKeys.list(filter),
    queryFn: ({ signal }) => fetchTests(filter, signal),
    staleTime: 30_000,
  });
}