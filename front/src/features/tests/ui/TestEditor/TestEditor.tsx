import type { FC } from "react";
import { useState } from "react";
import type {
  QuestionType,
  TestStatus,
} from "../../../../entities/test/model/types";
import { QuestionList } from "../QuestionList/QuestionList";
import { QuestionEditor } from "../QuestionEditor/QuestionEditor";
import { QuestionSettings } from "../QuestionSettings/QuestionSettings";
import { PreviewModal } from "../PreviewModal/PreviewModal";
import { TestSettingsModal } from "../TestSettingsModal/TestSettingsModal";
import styles from "./TestEditor.module.css";
import type {
  EditorDraft,
  EditorOption,
  EditorQuestion,
} from "../../model/editorTypes";

type TestEditorProps = {
  draft: EditorDraft;
  status: TestStatus;
  revision: number;
  isDirty: boolean;
  isSaving: boolean;
  saveError: string | null;
  selectedKey: string | null;
  onSelect: (key: string) => void;
  onBack: () => void;
  onSave: () => void;
  onMetaChange: (patch: Partial<EditorDraft["meta"]>) => void;
  onAddQuestion: (type: QuestionType) => void;
  onRemoveQuestion: (key: string) => void;
  onPatchQuestion: (key: string, patch: Partial<EditorQuestion>) => void;
  onChangeType: (key: string, type: QuestionType) => void;
  onMoveQuestion: (key: string, direction: -1 | 1) => void;
  onAddOption: (key: string) => void;
  onRemoveOption: (key: string, oKey: string) => void;
  onPatchOption: (
    key: string,
    oKey: string,
    patch: Partial<EditorOption>,
  ) => void;
  onSetCorrectOption: (key: string, oKey: string) => void;
};

export const TestEditor: FC<TestEditorProps> = (props) => {
  const {
    draft,
    status,
    revision,
    isDirty,
    isSaving,
    saveError,
    selectedKey,
    onSelect,
    onBack,
    onSave,
    onMetaChange,
    onAddQuestion,
    onRemoveQuestion,
    onPatchQuestion,
    onChangeType,
    onMoveQuestion,
    onAddOption,
    onRemoveOption,
    onPatchOption,
    onSetCorrectOption,
  } = props;

  const [previewOpen, setPreviewOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);

  const isPublished = status !== "draft";
  const selectedQuestion =
    draft.questions.find((q) => q.clientKey === selectedKey) ?? null;
  const selectedIndex = selectedQuestion
    ? draft.questions.findIndex(
        (q) => q.clientKey === selectedQuestion.clientKey,
      )
    : -1;

  const dirtyLabel = isSaving
    ? "Сохранение..."
    : saveError
      ? "Ошибка сохранения"
      : isDirty
        ? "Есть несохранённые изменения"
        : "Все изменения сохранены";

  const dirtyClass = isSaving
    ? styles.statusSaving
    : saveError
      ? styles.statusError
      : isDirty
        ? styles.statusDirty
        : styles.statusSaved;

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.headerLeft}>
          <button
            className={styles.backButton}
            onClick={onBack}
            title="К списку тестов"
          >
            ←
          </button>
          <div>
            <h1
              className={styles.title}
              onClick={() => setSettingsOpen(true)}
              title="Настройки теста"
            >
              {draft.meta.title || "(без названия)"}
            </h1>
            <div className={styles.meta}>
              <span className={styles.statusBadge}>Черновик</span>
              <span className={styles.revision}>v{revision}</span>
              <span className={`${styles.dirtyStatus} ${dirtyClass}`}>
                {dirtyLabel}
              </span>
            </div>
          </div>
        </div>
        <div className={styles.headerActions}>
          <button
            className={styles.secondaryButton}
            onClick={() => setPreviewOpen(true)}
          >
            Предпросмотр
          </button>
          <button
            className={styles.primaryButton}
            onClick={onSave}
            disabled={!isDirty || isSaving || isPublished}
          >
            {isSaving ? "Сохранение..." : "Сохранить"}
          </button>
        </div>
      </header>

      {isPublished && (
        <div className={styles.publishedBanner}>
          Тест опубликован. Редактирование недоступно.
        </div>
      )}

      <div className={styles.workspace}>
        <QuestionList
          questions={draft.questions}
          selectedKey={selectedKey}
          onSelect={onSelect}
          onAdd={onAddQuestion}
          onMove={onMoveQuestion}
          disabled={isPublished}
        />

        <main className={styles.center}>
          {selectedQuestion ? (
            <QuestionEditor
              question={selectedQuestion}
              index={selectedIndex}
              onChange={onPatchQuestion}
              onChangeType={onChangeType}
              onAddOption={onAddOption}
              onRemoveOption={onRemoveOption}
              onPatchOption={onPatchOption}
              onSetCorrectOption={onSetCorrectOption}
              disabled={isPublished}
            />
          ) : (
            <div className={styles.emptyCenter}>
              {draft.questions.length === 0
                ? "Добавьте первый вопрос, чтобы начать."
                : "Выберите вопрос слева."}
            </div>
          )}
        </main>

        <aside className={styles.rightPanel}>
          {selectedQuestion ? (
            <QuestionSettings
              question={selectedQuestion}
              index={selectedIndex}
              total={draft.questions.length}
              onChange={onPatchQuestion}
              onRemove={(key) => {
                if (window.confirm("Удалить вопрос?")) onRemoveQuestion(key);
              }}
              onMove={onMoveQuestion}
              disabled={isPublished}
            />
          ) : (
            <div className={styles.settingsPlaceholder}>Настройки вопроса</div>
          )}
        </aside>
      </div>

      {previewOpen && (
        <PreviewModal draft={draft} onClose={() => setPreviewOpen(false)} />
      )}
      {settingsOpen && (
        <TestSettingsModal
          meta={draft.meta}
          onChange={onMetaChange}
          onClose={() => setSettingsOpen(false)}
          disabled={isPublished}
        />
      )}
    </div>
  );
};
