export const analyticsQueryKeys = {
  all: ["analytics"] as const,
  recent: (limit: number) => ["analytics", "recent", limit] as const,
};