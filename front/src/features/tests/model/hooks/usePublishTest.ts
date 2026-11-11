import { useMutation, useQueryClient } from "@tanstack/react-query";
import { publishTest } from "../../api/publishTest";
import { testsQueryKeys } from "../queryKeys";
import { dashboardQueryKeys } from "../../../dashboard-overview/model/queryKeys";
import type { PublicationResponse } from "../../../../entities/publication";

export function usePublishTest(testId: string) {
  const queryClient = useQueryClient();

  return useMutation<PublicationResponse, Error, number>({
    mutationFn: (expectedRevision) => publishTest(testId, expectedRevision),
    onSuccess: async () => {
      await queryClient.refetchQueries({ queryKey: testsQueryKeys.detail(testId) });
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.overview });
      queryClient.invalidateQueries({ queryKey: testsQueryKeys.all });
    },
  });
}