import type { ApiError } from "./types";

const API_BASE = import.meta.env.VITE_API_URL ?? "/api";

let csrfToken: string | null = null


export function setCsrfToken(token: string | null) {
  csrfToken = token;
}

export function getCsrfToken() {
  return csrfToken;
}export class HttpError extends Error {
  readonly code: string;
   readonly status: number;
  constructor(
     status: number,
     code: string,
    message: string,
  ) {
    super(message);
    this.name = "HttpError";
    this.status = status;
    this.code = code;
  }
 
}

type RequestOptions = {
  method?: "GET" | "POST" | "PUT"  | "DELETE";
  body?: unknown;
  signal?: AbortSignal;
};

export async function apiFetch<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { method = "GET", body, signal } = options;

  const headers: HeadersInit = {
    Accept: "application/json",
  };

  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  // CSRF нужен только для небезопасных методов
  const isUnsafe = method !== "GET";
  if (isUnsafe && csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    credentials: "include",
    body: body !== undefined ? JSON.stringify(body) : undefined,
    signal,
  });

  // 204 No Content — тела нет
  if (res.status === 204) {
    return undefined as T;
  }

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    const apiErr = data as ApiError | null;
    throw new HttpError(
      res.status,
      apiErr?.error?.code ?? "UNKNOWN",
      apiErr?.error?.message ?? `Request failed: ${res.status}`,
    );
  }

  return data as T;
}