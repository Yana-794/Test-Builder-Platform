import type { FC } from "react";
import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router";

import { TestEditor } from "../../features/tests/ui/TestEditor/TestEditor";
import { ConflictModal } from "../../features/tests/ui/ConflictModal/ConflictModal";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";
import { ApiError } from "../../shared/api/http";
import { useTestEditor } from "../../features/tests/model/hooks/useTestEditor";
import {
  isRevisionConflict,
  useSaveDraft,
} from "../../features/tests/model/hooks/useSaveDraft";
import { buildDraftPayload } from "../../features/tests/model/buildDraftPayload";
import { PublishModal } from "../../features/tests/ui/PublishModal/PublishModal";

export const TestEditorPage: FC = () => {
  const { testId = "" } = useParams<{ testId: string }>();
  const navigate = useNavigate();

  const editor = useTestEditor(testId);
  const save = useSaveDraft(testId);

  const [selectedKey, setSelectedKey] = useState<string | null>(null);
  const [conflictOpen, setConflictOpen] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [publishOpen, setPublishOpen] = useState(false);

  // Выбор первого вопроса при первой загрузке
  useEffect(() => {
    if (!editor.draft) return;
    if (selectedKey) return;
    if (editor.draft.questions.length > 0) {
      setSelectedKey(editor.draft.questions[0].clientKey);
    }
  }, [editor.draft, selectedKey]);

  useEffect(() => {
    if (!editor.isDirty) return;
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [editor.isDirty]);

  const isUnauthorized =
    editor.error instanceof ApiError && editor.error.status === 401;
  const isNotFound =
    editor.error instanceof ApiError && editor.error.status === 404;

  useEffect(() => {
    if (isUnauthorized) navigate("/login", { replace: true });
  }, [isUnauthorized, navigate]);

  const handleBack = () => {
    if (
      editor.isDirty &&
      !window.confirm("Есть несохранённые изменения. Покинуть редактор?")
    ) {
      return;
    }
    navigate("/app/tests");
  };

  const handleSave = async () => {
    if (editor.revision === null) return;
    setSaveError(null);
    try {
      const payload = buildDraftPayload(editor.draft!, editor.revision);
      const saved = await save.mutateAsync(payload);
      editor.applySavedTest(saved);
    } catch (err) {
      if (isRevisionConflict(err)) {
        setConflictOpen(true);
        return;
      }
      if (err instanceof ApiError && err.status === 401) {
        navigate("/login", { replace: true });
        return;
      }
      setSaveError(
        "Не удалось сохранить. Проверьте поля и попробуйте ещё раз.",
      );
    }
  };

  const handleConflictRefresh = async () => {
    setConflictOpen(false);
    await editor.refetch();
  };

  if (editor.isLoading) {
    return <div style={{ padding: 24 }}>Загрузка...</div>;
  }

  if (isNotFound) {
    return (
      <ErrorState
        message="Тест не найден или недоступен"
        retryLabel="К списку тестов"
        onRetry={() => navigate("/app/tests")}
      />
    );
  }

  if (editor.isError || !editor.draft || !editor.serverTest) {
    return (
      <ErrorState
        message="Не удалось загрузить тест"
        onRetry={() => editor.refetch()}
      />
    );
  }

  return (
    <>
      <TestEditor
        draft={editor.draft}
        status={editor.serverTest.status}
        revision={editor.serverTest.revision}
        isDirty={editor.isDirty}
        isSaving={save.isPending}
        saveError={saveError}
        selectedKey={selectedKey}
        onSelect={setSelectedKey}
        onBack={handleBack}
        onPublishClick={() => setPublishOpen(true)}
        onSave={handleSave}
        onMetaChange={editor.setMeta}
        onAddQuestion={(type) => {
          const key = editor.addQuestion(type);
          setSelectedKey(key);
        }}
        onRemoveQuestion={(key) => {
          const idx =
            editor.draft?.questions.findIndex((q) => q.clientKey === key) ?? -1;
          editor.removeQuestion(key);
          if (selectedKey !== key) return;

          const remaining =
            editor.draft?.questions.filter((q) => q.clientKey !== key) ?? [];
          if (remaining.length === 0) {
            setSelectedKey(null);
          } else {
            const nextIdx = Math.min(Math.max(idx, 0), remaining.length - 1);
            setSelectedKey(remaining[nextIdx].clientKey);
          }
        }}
        onPatchQuestion={editor.patchQuestion}
        onChangeType={editor.changeQuestionType}
        onMoveQuestion={editor.moveQuestion}
        onAddOption={editor.addOption}
        onRemoveOption={editor.removeOption}
        onPatchOption={editor.patchOption}
        onSetCorrectOption={editor.setCorrectOption}
      />
      {publishOpen && editor.draft && editor.serverTest && (
        <PublishModal
          testId={testId}
          testTitle={editor.draft.meta.title}
          testDescription={editor.draft.meta.description}
          revision={editor.serverTest.revision}
          onClose={() => setPublishOpen(false)}
        />
      )}

      {conflictOpen && (
        <ConflictModal
          onRefresh={handleConflictRefresh}
          onStay={() => setConflictOpen(false)}
        />
      )}
    </>
  );
};
