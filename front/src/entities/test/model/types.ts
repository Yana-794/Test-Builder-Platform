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
// ─────────────────────────────────────────────────────────────
// TBP-02: типы для создания и редактирования теста
// ─────────────────────────────────────────────────────────────

export type QuestionType = "single_choice" | "text";

export type AnswerOption ={
  id: string;
  text: string;
  isCorrect: boolean;
  position: number;
}

export type Question = {
  id: string;
  prompt: string;
  type: QuestionType;
  points: number;
  position: number;
  options?: AnswerOption[];
  correctTextAnswer?: string;
}

// Базовые поля теста (без вопросов) — ответ POST /tests,
// а также часть ответа GET /tests/:id.
export type Test = {
  id: string;
  title: string;
  description: string;
  subject: string | null;
  status: TestStatus;
  revision: number;
  createdAt: string;
  updatedAt: string;
}

// Ответ GET /tests/:id и PUT /tests/:id/draft
export type TestEditor = Test & {
  questions: Question[];
};

export type CreateTestPayload = {
  title: string;
  description: string;
  subject: string | null;
}

// Для сохранения черновика: id === null у новых вопросов/опций.
export type DraftQuestion = {
  id: string | null;
  prompt: string;
  type: QuestionType;
  points: number;
  options?: DraftOption[];
  correctTextAnswer?: string | null;
}

export type DraftOption = {
  id: string | null;
  text: string;
  isCorrect: boolean;
}

export type SaveDraftPayload = {
  expectedRevision: number;
  test: {
    title: string;
    description: string;
    subject: string | null;
  };
  questions: DraftQuestion[];
}