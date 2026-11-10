import { useMutation } from "@tanstack/react-query";
import { validateTest } from "../../api/validateTest";
import type { ValidateResponse } from "../../../../entities/test/model/types";

export function useValidateTest(testId: string) {
  return useMutation<ValidateResponse, Error, void>({
    mutationFn: () => validateTest(testId),
    retry: false,
  });
}