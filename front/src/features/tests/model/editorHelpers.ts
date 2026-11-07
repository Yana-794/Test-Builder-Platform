import type { Question } from "../../../entities/test/model/types";
import type {
  EditorDraft,
  EditorQuestion,
  EditorOption,
} from "./editorTypes";

export const newClientKey = (): string =>
  typeof crypto !== "undefined" && "randomUUID" in crypto
    ? crypto.randomUUID()
    : `tmp-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

export function toEditorQuestion(q: Question): EditorQuestion {
  return {
    clientKey: newClientKey(),
    serverId: q.id,
    prompt: q.prompt,
    type: q.type,
    points: q.points,
    correctTextAnswer: q.correctTextAnswer ?? null,
    options: (q.options ?? []).map((o) => ({
      clientKey: newClientKey(),
      serverId: o.id,
      text: o.text,
      isCorrect: o.isCorrect,
    })),
  };
}

export function newQuestion(type: "single_choice" | "text"): EditorQuestion {
  const base: EditorQuestion = {
    clientKey: newClientKey(),
    serverId: null,
    prompt: "",
    type,
    points: 1,
    options: [],
    correctTextAnswer: null,
  };

  if (type === "single_choice") {
    base.options = [
      { clientKey: newClientKey(), serverId: null, text: "", isCorrect: false },
      { clientKey: newClientKey(), serverId: null, text: "", isCorrect: false },
    ];
  }
  return base;
}

export function newOption(): EditorOption {
  return {
    clientKey: newClientKey(),
    serverId: null,
    text: "",
    isCorrect: false,
  };
}

// deep-equal для dirty-детектора. Сравнивает только то, что уходит на сервер, игнорируя clientKey
export function draftToServerPayload(draft: EditorDraft): string {
  return JSON.stringify({
    title: draft.meta.title,
    description: draft.meta.description,
    subject: draft.meta.subject,
    questions: draft.questions.map((q) => ({
      id: q.serverId,
      prompt: q.prompt,
      type: q.type,
      points: q.points,
      options:
        q.type === "single_choice"
          ? q.options.map((o) => ({
              id: o.serverId,
              text: o.text,
              isCorrect: o.isCorrect,
            }))
          : undefined,
      correctTextAnswer: q.type === "text" ? q.correctTextAnswer : undefined,
    })),
  });
}

//  помощники для апдейтов
export function updateQuestion(
  draft: EditorDraft,
  clientKey: string,
  patch: Partial<EditorQuestion>,
): EditorDraft {
  return {
    ...draft,
    questions: draft.questions.map((q) =>
      q.clientKey === clientKey ? { ...q, ...patch } : q,
    ),
  };
}

export function updateOption(
  draft: EditorDraft,
  qKey: string,
  oKey: string,
  patch: Partial<EditorOption>,
): EditorDraft {
  return {
    ...draft,
    questions: draft.questions.map((q) =>
      q.clientKey !== qKey
        ? q
        : {
            ...q,
            options: q.options.map((o) =>
              o.clientKey === oKey ? { ...o, ...patch } : o,
            ),
          },
    ),
  };
}