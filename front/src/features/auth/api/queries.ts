import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch, HttpError, setCsrfToken } from "./client";
import { authKeys } from "./queryKeys";
import type {
  AuthResponse,
  LoginPayload,
  RegisterPayload,
  User,
} from "./types";

export function useMe() {
  return useQuery<AuthResponse | null>({
    queryKey: authKeys.me(),
    queryFn: async () => {
      try {
        const data = await apiFetch<AuthResponse>("/auth/me");
        setCsrfToken(data.csrf_token);
        return data;
      } catch (err) {
        // 401
        if (err instanceof HttpError && err.status === 401) {
          setCsrfToken(null);
          return null;
        }
        throw err;
      }
    },
    staleTime: 5 * 60 * 1000, // 5 минут
    retry: false, // 401
  });
}

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: LoginPayload) =>
      apiFetch<AuthResponse>("/auth/login", {
        method: "POST",
        body: payload,
      }),
    onSuccess: (data) => {
      setCsrfToken(data.csrf_token);
      queryClient.setQueryData(authKeys.me(), data);
    },
  });
}

export function useRegister() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: RegisterPayload) =>
      apiFetch<{ user: User }>("/auth/register", {
        method: "POST",
        body: payload,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.all });
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => apiFetch<void>("/auth/logout", { method: "POST" }),
    onSuccess: () => {
      setCsrfToken(null);
      queryClient.setQueryData(authKeys.me(), null);
      queryClient.removeQueries({ queryKey: authKeys.all });
    },
    onError: () => {
      setCsrfToken(null);
      queryClient.setQueryData(authKeys.me(), null);
    },
  });
}
