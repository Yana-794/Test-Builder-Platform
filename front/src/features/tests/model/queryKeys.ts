export const testsQueryKeys = {
  all: ["tests"] as const,
  detail: (testId: string) => ["tests", "detail", testId] as const,
};