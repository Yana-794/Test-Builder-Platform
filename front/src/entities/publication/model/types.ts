import type { QuestionType } from "../../test";

export type ValidateResponse = {
  valid: boolean;
  errors: PublicationFieldDetail[];
  summary: {
    questionCount: number;
    maxScore: number;
  };
};
export type PublicationFieldDetail = {
  field: string;
  code: string;
  message: string;
};
export type Publication = {
  id: string;
  version: number;
  publicToken: string;
  publicUrl: string;
  publishedAt: string;
};

export type PublicationResponse = {
  publication: Publication;
};

export type PublicOption = {
  id: string;
  text: string;
};

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