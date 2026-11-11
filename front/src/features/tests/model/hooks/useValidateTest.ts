import { useMutation } from "@tanstack/react-query";
import { validateTest } from "../../api/validateTest";
import type {  } from "../../../../entities/test/model/types";
import type { ValidateResponse } from "../../../../entities/publication";

export function useValidateTest(testId: string) {
  return useMutation<ValidateResponse, Error, void>({
    mutationFn: () => validateTest(testId),
    retry: false,
  });
}