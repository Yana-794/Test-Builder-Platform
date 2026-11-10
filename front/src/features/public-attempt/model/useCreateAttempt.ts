import { useMutation } from "@tanstack/react-query";
import { createAttempt } from "../api/createAttempt";

export function useCreateAttempt(token: string) {
  return useMutation({
    mutationFn: (studentName: string) => createAttempt(token, studentName),
  });
}