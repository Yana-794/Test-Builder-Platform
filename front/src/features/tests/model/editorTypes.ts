import type { QuestionType } from "../../../entities/test/model/types";

export type EditorQuestion = {
  clientKey: string;
  serverId: string | null;
  prompt: string;
  type: QuestionType;
  points: number;
  options: EditorOption[];
  correctTextAnswer: string | null;
};

export type EditorOption = {
  clientKey: string;
  serverId: string | null;
  text: string;
  isCorrect: boolean;
};

export type EditorTestMeta = {
  title: string;
  description: string;
  subject: string | null;
};

export type EditorDraft = {
  meta: EditorTestMeta;
  questions: EditorQuestion[];
};