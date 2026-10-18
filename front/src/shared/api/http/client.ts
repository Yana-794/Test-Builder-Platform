import { API_BASE_URL } from "../../config";
import { ApiError } from "./ApiError";

/** Таймаут запроса по умолчанию (мс) */
const DEFAULT_TIMEOUT_MS = 15_000;

/** Методы, для которых бэк требует X-CSRF-Token */
const UNSAFE_METHODS = new Set(["POST", "PUT", "PATCH", "DELETE"]);

// CSRF 
let csrfToken: string | null = null;
export function setCsrfToken(token: string | null): void {
  csrfToken = token;
}
export function getCsrfToken(): string | null {
  return csrfToken;
}

//  401 handler 
let unauthorizedHandler: (() => void) | null = null;
export function setUnauthorizedHandler(handler: (() => void) | null): void {
  unauthorizedHandler = handler;
}

// Алиас, чтобы не ломать импорты `HttpError` в queries.ts
export { ApiError, ApiError as HttpError };

//  Options 
export type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  query?: Record<string, string | number | boolean | undefined | null>;
  body?: unknown;
  signal?: AbortSignal;
  timeoutMs?: number;
};

function buildUrl(path: string, query?: RequestOptions["query"]): string {
  const base = API_BASE_URL.endsWith("/")
    ? API_BASE_URL.slice(0, -1)
    : API_BASE_URL;
  const url = `${base}${path.startsWith("/") ? path : `/${path}`}`;
  if (!query) return url;
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== null) params.set(key, String(value));
  }
  const qs = params.toString();
  return qs ? `${url}?${qs}` : url;
}

/** Разбор тела ответа (один раз) */
async function readBody(res: Response): Promise<unknown> {
  if (res.status === 204) return undefined;
  const text = await res.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

/** Достаём { error: { code, message } } из Go-бэка */
function parseBackendError(status: number, data: unknown): ApiError {
  let code: string | undefined;
  let message = `Запрос завершился ошибкой ${status}`;

  if (data && typeof data === "object" && "error" in data) {
    const err = (data as { error?: { code?: unknown; message?: unknown } })
      .error;
    if (err && typeof err === "object") {
      if (typeof err.code === "string") code = err.code;
      if (typeof err.message === "string") message = err.message;
    }
  }

  return new ApiError(message, status, { code, details: data });
}

// ---------- apiFetch ----------
export async function apiFetch<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const {
    method = "GET",
    query,
    body,
    signal,
    timeoutMs = DEFAULT_TIMEOUT_MS,
  } = options;

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);
  if (signal) {
    if (signal.aborted) controller.abort();
    else
      signal.addEventListener("abort", () => controller.abort(), {
        once: true,
      });
  }

  const isFormData = body instanceof FormData;
  const headers: Record<string, string> = { Accept: "application/json" };

  if (body !== undefined && !isFormData) {
    headers["Content-Type"] = "application/json";
  }

  // CSRF только для небезопасных методов и только если токен уже получен
  if (UNSAFE_METHODS.has(method) && csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }

  let response: Response;
  try {
    response = await fetch(buildUrl(path, query), {
      method,
      headers,
      credentials: "include",
      body: isFormData
        ? (body as FormData)
        : body !== undefined
          ? JSON.stringify(body)
          : undefined,
      signal: controller.signal,
    });
  } catch (err) {
    const aborted = err instanceof DOMException && err.name === "AbortError";
    throw new ApiError(aborted ? "Запрос прерван" : "Ошибка сети", 0, {
      code: aborted ? "ABORTED" : "NETWORK_ERROR",
      details: err,
    });
  } finally {
    clearTimeout(timeout);
  }

  const data = await readBody(response);

  if (!response.ok) {
    if (response.status === 401) unauthorizedHandler?.();
    throw parseBackendError(response.status, data);
  }

  return data as T;
}

export const httpClient = {
  get: <T>(path: string, options?: Omit<RequestOptions, "method" | "body">) =>  apiFetch<T>(path, { ...options, method: "GET" }),
  post: <T>(path: string, options?: Omit<RequestOptions, "method">) =>  apiFetch<T>(path, { ...options, method: "POST" }),
  put: <T>(path: string, options?: Omit<RequestOptions, "method">) =>  apiFetch<T>(path, { ...options, method: "PUT" }),
  delete: <T>(path: string, options?: Omit<RequestOptions, "method">) => apiFetch<T>(path, { ...options, method: "DELETE" }),
};
