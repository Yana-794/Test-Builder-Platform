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

export type QuestionType = "single_choice" | "text";

export type AnswerOption = {
  id: string;
  text: string;
  isCorrect: boolean;
  position: number;
};

export type Question = {
  id: string;
  prompt: string;
  type: QuestionType;
  points: number;
  position: number;
  options?: AnswerOption[];
  correctTextAnswer?: string;
};

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
};

// Ответ GET /tests/:id и PUT /tests/:id/draft
export type TestEditor = Test & {
  questions: Question[];
};

export type CreateTestPayload = {
  title: string;
  description: string;
  subject: string | null;
};

// Для сохранения черновика: id === null у новых вопросов/опций.
export type DraftQuestion = {
  id: string | null;
  prompt: string;
  type: QuestionType;
  points: number;
  options?: DraftOption[];
  correctTextAnswer?: string | null;
};

export type DraftOption = {
  id: string | null;
  text: string;
  isCorrect: boolean;
};

export type SaveDraftPayload = {
  expectedRevision: number;
  test: {
    title: string;
    description: string;
    subject: string | null;
  };
  questions: DraftQuestion[];
};

export type PublicationFieldDetail = {
  field: string;
  code: string;
  message: string;
};

export type ValidateResponse = {
  valid: boolean;
  errors: PublicationFieldDetail[];
  summary: {
    questionCount: number;
    maxScore: number;
  };
};

export type Publication = {
  id: string;
  version: number;
  publicToken: string;
  publicUrl: string;
  publishedAt: string;
};

export type TestFilter = "all" | "draft" | "published" | "archived";

export type TestListItem = {
  id: string;
  title: string;
  status: TestStatus;
  revision: number;
  updatedAt: string;
};

export type TestListResponse = {
  tests: TestListItem[];
  filter: string;
  total: number;
};

export type PublicOption = { id: string; text: string };

export type PublicQuestion = {
  id: string;
  prompt: string;
  type: QuestionType;
  position: number;
  options?: PublicOption[];
};

export type PublicTest = {
  title: string;
  description: string;
  questionCount: number;
  questions: PublicQuestion[];
};

export type AttemptShort = {
  id: string;
  status: "in_progress" | "completed";
  startedAt: string;
};

export type CreateAttemptResponse = { attempt: AttemptShort };

export type AnswerDTO = {
  questionId: string;
  selectedOptionId?: string;
  answerText?: string;
};

export type AttemptStateResponse = {
  attempt: AttemptShort;
  test: PublicTest;
  answers: AnswerDTO[];
};

export type AttemptResult = {
  id: string;
  status: "in_progress" | "completed";
  score: number;
  maxScore: number;
  percentage: number;
  durationSeconds: number;
  completedAt: string;
};

export type AttemptResultResponse = { attempt: AttemptResult };

export type SaveAnswersPayload = { answers: AnswerDTO[] };

export type PublicationResponse = {
  publication: Publication;
};
