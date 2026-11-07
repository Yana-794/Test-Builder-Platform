import type {
  DraftOption,
  DraftQuestion,
  SaveDraftPayload,
} from "../../../entities/test/model/types";
import type { EditorDraft } from "./editorTypes";

export function buildDraftPayload(
  draft: EditorDraft,
  expectedRevision: number,
): SaveDraftPayload {
  const questions: DraftQuestion[] = draft.questions.map((q) => {
    const base: DraftQuestion = {
      id: q.serverId,
      prompt: q.prompt,
      type: q.type,
      points: q.points,
    };

    if (q.type === "single_choice") {
      const options: DraftOption[] = q.options.map((o) => ({
        id: o.serverId,
        text: o.text,
        isCorrect: o.isCorrect,
      }));
      base.options = options;
    } else {
      base.correctTextAnswer = q.correctTextAnswer;
    }
    return base;
  });

  return {
    expectedRevision,
    test: {
      title: draft.meta.title,
      description: draft.meta.description,
      subject: draft.meta.subject,
    },
    questions,
  };
}