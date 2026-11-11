export type AnalyticsRecentAttempt = {
  attemptId: string;
  studentName: string;
  testId: string;
  testTitle: string;
  status: "in_progress" | "completed";
  score?: number;
  maxScore?: number;
  percent?: number;
  durationSeconds?: number;
  startedAt: string;
  completedAt?: string;
};