export const testsQueryKeys = {
  all: ["tests"] as const,
  list: (filter: string) => ["tests", "list", filter] as const,
  detail: (testId: string) => ["tests", "detail", testId] as const,
};