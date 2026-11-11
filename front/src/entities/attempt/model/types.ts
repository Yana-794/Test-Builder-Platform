import type { PublicTest } from "../../publication";

export type AttemptShort = {
  id: string;
  status: "in_progress" | "completed";
  startedAt: string;
};

export type CreateAttemptResponse = {
  attempt: AttemptShort;
};

export type AnswerDTO = {
  questionId: string;
  selectedOptionId?: string;
  answerText?: string;
};

export type SaveAnswersPayload = {
  answers: AnswerDTO[];
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

export type AttemptResultResponse = {
  attempt: AttemptResult;
};