import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { getTest } from "../../api/getTest";
import { testsQueryKeys } from "../queryKeys";
import type { TestEditor } from "../../../../entities/test/model/types";
import type { EditorDraft, EditorQuestion, EditorOption } from "../editorTypes";
import {
  draftToServerPayload,
  newOption,
  newQuestion,
  toEditorQuestion,
  updateOption as updateOptionHelper,
  updateQuestion as updateQuestionHelper,
} from "../editorHelpers";

function serverToDraft(test: TestEditor): EditorDraft {
  return {
    meta: {
      title: test.title,
      description: test.description,
      subject: test.subject,
    },
    questions: [...test.questions]
      .sort((a, b) => a.position - b.position)
      .map(toEditorQuestion),
  };
}

export function useTestEditor(testId: string) {
  const query = useQuery({
    queryKey: testsQueryKeys.detail(testId),
    queryFn: ({ signal }) => getTest(testId, signal),
    retry: false,
    staleTime: Infinity,
  });

  const serverTest = query.data?.test ?? null;
  const [draft, setDraft] = useState<EditorDraft | null>(null);
  const [savedSnapshot, setSavedSnapshot] = useState<string | null>(null);

  useEffect(() => {
    if (!serverTest) return;
    const initial = serverToDraft(serverTest);
    setDraft(initial);
    setSavedSnapshot(draftToServerPayload(initial));
  }, [serverTest?.id, serverTest?.revision]); // не переинициализировать при каждом рендере

  const isDirty = useMemo(() => {
    if (!draft || savedSnapshot === null) return false;
    return draftToServerPayload(draft) !== savedSnapshot;
  }, [draft, savedSnapshot]);

  const setMeta = (patch: Partial<EditorDraft["meta"]>) => {
    setDraft((d) => (d ? { ...d, meta: { ...d.meta, ...patch } } : d));
  };

  const addQuestion = (type: "single_choice" | "text"): string => {
    const q = newQuestion(type);
    setDraft((d) => (d ? { ...d, questions: [...d.questions, q] } : d));
    return q.clientKey;
  };

  const removeQuestion = (clientKey: string) => {
    setDraft((d) =>
      d
        ? {
            ...d,
            questions: d.questions.filter((q) => q.clientKey !== clientKey),
          }
        : d,
    );
  };

  const patchQuestion = (clientKey: string, patch: Partial<EditorQuestion>) => {
    setDraft((d) => (d ? updateQuestionHelper(d, clientKey, patch) : d));
  };

  const changeQuestionType = (
    clientKey: string,
    type: "single_choice" | "text",
  ) => {
    setDraft((d) => {
      if (!d) return d;
      return {
        ...d,
        questions: d.questions.map((q) => {
          if (q.clientKey !== clientKey || q.type === type) return q;
          return {
            ...q,
            type,
            // при смене типа сбрасывает несовместимые поля
            correctTextAnswer: type === "text" ? q.correctTextAnswer : null,
            options:
              type === "single_choice"
                ? q.options.length >= 2
                  ? q.options
                  : [newOption(), newOption()]
                : [],
          };
        }),
      };
    });
  };

  const addOption = (clientKey: string) => {
    setDraft((d) => {
      if (!d) return d;
      return {
        ...d,
        questions: d.questions.map((q) =>
          q.clientKey !== clientKey
            ? q
            : { ...q, options: [...q.options, newOption()] },
        ),
      };
    });
  };

  const removeOption = (clientKey: string, optionClientKey: string) => {
    setDraft((d) => {
      if (!d) return d;
      return {
        ...d,
        questions: d.questions.map((q) =>
          q.clientKey !== clientKey
            ? q
            : {
                ...q,
                options: q.options.filter(
                  (o) => o.clientKey !== optionClientKey,
                ),
              },
        ),
      };
    });
  };

  const patchOption = (
    clientKey: string,
    optionClientKey: string,
    patch: Partial<EditorOption>,
  ) => {
    setDraft((d) =>
      d ? updateOptionHelper(d, clientKey, optionClientKey, patch) : d,
    );
  };

  const setCorrectOption = (clientKey: string, optionClientKey: string) => {
    setDraft((d) => {
      if (!d) return d;
      return {
        ...d,
        questions: d.questions.map((q) =>
          q.clientKey !== clientKey
            ? q
            : {
                ...q,
                options: q.options.map((o) => ({
                  ...o,
                  isCorrect: o.clientKey === optionClientKey,
                })),
              },
        ),
      };
    });
  };

  const moveQuestion = (clientKey: string, direction: -1 | 1) => {
    setDraft((d) => {
      if (!d) return d;
      const idx = d.questions.findIndex((q) => q.clientKey === clientKey);
      if (idx < 0) return d;
      const target = idx + direction;
      if (target < 0 || target >= d.questions.length) return d;
      const next = [...d.questions];
      [next[idx], next[target]] = [next[target], next[idx]];
      return { ...d, questions: next };
    });
  };

  // После успешного сохранения
  // Обновляет серверные id/revision в локальном состоянии и сбрасывает dirty
  const applySavedTest = (saved: TestEditor) => {
    const next = serverToDraft(saved);
    setDraft(next);
    setSavedSnapshot(draftToServerPayload(next));
  };

  return {
    isLoading: query.isLoading,
    isError: query.isError,
    error: query.error,
    refetch: query.refetch,
    serverTest,
    revision: serverTest?.revision ?? null,
    status: serverTest?.status ?? null,

    // локальное состояние
    draft,
    isDirty,

    // действия
    setMeta,
    addQuestion,
    removeQuestion,
    patchQuestion,
    changeQuestionType,
    addOption,
    removeOption,
    patchOption,
    setCorrectOption,
    moveQuestion,
    applySavedTest,
  };
}
