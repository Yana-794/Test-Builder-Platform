import type { PublicTest } from "../../../entities/publication";
import { httpClient } from "../../../shared/api/http";

export const fetchPublicTest = (token: string, signal?: AbortSignal) =>
  httpClient.get<{ test: PublicTest }>(`/public/tests/${token}`, { signal });