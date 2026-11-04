export type TestStatus = "draft" | "published";

export type RecentTest = {
  id: string;
  title: string;
  status: TestStatus;
  createdAt: string;
};

export type DashboardStats = {
  totalTests: number;
  draftTests: number;
  publishedTests: number;
};

export type DashboardOverview = {
  stats: DashboardStats;
  recentTests: RecentTest[];
};