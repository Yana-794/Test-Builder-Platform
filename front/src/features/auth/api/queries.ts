import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getCsrfToken } from "../../../shared/api/http/client";
import { httpClient, setCsrfToken } from "../../../shared/api/http/client";
import { ApiError } from "../../../shared/api/http";
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
        const data = await httpClient.get<AuthResponse>("/auth/me");
        setCsrfToken(data.csrf_token);
        return data;
      } catch (err) {
        if (err instanceof ApiError && err.status === 401) {
          setCsrfToken(null);
          return null;
        }
        throw err;
      }
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: LoginPayload) =>
      httpClient.post<AuthResponse>("/auth/login", { body: payload }),
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
      httpClient.post<{ user: User }>("/auth/register", { body: payload }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.all });
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => {
      if (!getCsrfToken()) {
        const cached = queryClient.getQueryData<AuthResponse>(authKeys.me());
        if (cached?.csrf_token) setCsrfToken(cached.csrf_token);
      }
      return httpClient.post<void>("/auth/logout");
    },
    onSettled: () => {
      setCsrfToken(null);
      queryClient.setQueryData(authKeys.me(), null);
      queryClient.removeQueries({ queryKey: authKeys.all });
    },
  });
}