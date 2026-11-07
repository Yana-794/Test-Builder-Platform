import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createTest } from "../../api/createTest";
import type { CreateTestPayload, Test } from "../../../../entities/test/model/types";
import { dashboardQueryKeys } from "../../../dashboard-overview/model/queryKeys";

export function useCreateTest() {
  const queryClient = useQueryClient();

  return useMutation<Test, Error, CreateTestPayload>({
    mutationFn: async (payload) => {
      const resp = await createTest(payload);
      return resp.test;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: dashboardQueryKeys.overview });
    },
  });
}