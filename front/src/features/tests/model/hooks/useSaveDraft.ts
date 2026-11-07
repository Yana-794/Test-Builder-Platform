import { useMutation, useQueryClient } from "@tanstack/react-query";
import { saveDraft } from "../../api/saveDraft";
import { testsQueryKeys } from "../queryKeys";
import { dashboardQueryKeys } from "../../../dashboard-overview/model/queryKeys";
import type { TestEditor, SaveDraftPayload } from "../../../../entities/test/model/types";
import { ApiError } from "../../../../shared/api/http";

export function useSaveDraft(testId: string) {
  const queryClient = useQueryClient();

  return useMutation<TestEditor, Error, SaveDraftPayload>({
    mutationFn: async (payload) => {
      const resp = await saveDraft(testId, payload);
      return resp.test;
    },
    onSuccess: (test) => {
      queryClient.setQueryData(testsQueryKeys.detail(testId), { test });
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.overview });
      queryClient.invalidateQueries({ queryKey: testsQueryKeys.all });
    },
  });
}

export function isRevisionConflict(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    error.status === 409 &&
    error.message.toLowerCase().includes("draft")
  );
}