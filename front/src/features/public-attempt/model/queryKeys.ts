export const attemptQueryKeys = {
  publicTest: (token: string) => ["public", "test", token] as const,
  state: (id: string) => ["public", "attempt", id] as const,
  result: (id: string) => ["public", "attempt", id, "result"] as const,
};