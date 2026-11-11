import type { FC } from "react";
import { useState } from "react";
import {
  AlertCircle,
  ArrowLeft,
  CheckCircle2,
  Eye,
  FileQuestion,
  Loader2,
  Save,
  Send,
  Settings2,
} from "lucide-react";
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
  onPublishClick: () => void;
  onPatchOption: (
    key: string,
    oKey: string,
    patch: Partial<EditorOption>,
  ) => void;
  onSetCorrectOption: (key: string, oKey: string) => void;
};

const STATUS_LABEL: Record<TestStatus, string> = {
  draft: "Черновик",
  published: "Опубликован",
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
    onPublishClick,
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
            type="button"
            className={styles.backButton}
            onClick={onBack}
            title="К списку тестов"
            aria-label="Назад к списку тестов"
          >
            <ArrowLeft size={18} strokeWidth={2.2} />
          </button>

          <div className={styles.headerTitleBlock}>
            <button
              type="button"
              className={styles.titleButton}
              onClick={() => setSettingsOpen(true)}
              title="Настройки теста"
            >
              <span className={styles.title}>
                {draft.meta.title || "(без названия)"}
              </span>
              <Settings2
                size={14}
                strokeWidth={2.2}
                className={styles.titleIcon}
                aria-hidden="true"
              />
            </button>

            <div className={styles.meta}>
              <span
                className={`${styles.statusBadge} ${styles[`statusBadge_${status}`] ?? ""}`}
              >
                {STATUS_LABEL[status] ?? status}
              </span>
              <span className={styles.revision}>v{revision}</span>
              <span
                className={`${styles.dirtyStatus} ${dirtyClass}`}
                role="status"
                aria-live="polite"
              >
                <span className={styles.dirtyDot} aria-hidden="true" />
                {dirtyLabel}
              </span>
            </div>
          </div>
        </div>

        <div className={styles.headerActions}>
          {status === "draft" && (
            <button
              type="button"
              className={styles.publishButton}
              onClick={onPublishClick}
              disabled={isDirty || isSaving}
              title={
                isDirty
                  ? "Сохраните изменения перед публикацией"
                  : "Опубликовать тест"
              }
            >
              <Send size={15} strokeWidth={2.2} />
              <span>Опубликовать</span>
            </button>
          )}

          <button
            type="button"
            className={styles.secondaryButton}
            onClick={() => setPreviewOpen(true)}
          >
            <Eye size={15} strokeWidth={2.2} />
            <span>Предпросмотр</span>
          </button>

          <button
            type="button"
            className={styles.primaryButton}
            onClick={onSave}
            disabled={!isDirty || isSaving || isPublished}
          >
            {isSaving ? (
              <Loader2 size={15} strokeWidth={2.4} className={styles.spin} />
            ) : saveError ? (
              <AlertCircle size={15} strokeWidth={2.2} />
            ) : (
              <Save size={15} strokeWidth={2.2} />
            )}
            <span>{isSaving ? "Сохранение..." : "Сохранить"}</span>
          </button>
        </div>
      </header>

      {isPublished && (
        <div className={styles.publishedBanner} role="status">
          <CheckCircle2 size={16} strokeWidth={2.2} aria-hidden="true" />
          <span>Тест опубликован. Редактирование недоступно.</span>
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
              <span className={styles.emptyIcon} aria-hidden="true">
                <FileQuestion size={24} strokeWidth={2} />
              </span>
              <p className={styles.emptyTitle}>
                {draft.questions.length === 0
                  ? "Добавьте первый вопрос"
                  : "Вопрос не выбран"}
              </p>
              <p className={styles.emptyText}>
                {draft.questions.length === 0
                  ? "Чтобы начать, создайте вопрос слева."
                  : "Выберите вопрос слева, чтобы отредактировать его."}
              </p>
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
            <div className={styles.settingsPlaceholder}>
              Настройки появятся, когда вы выберете вопрос.
            </div>
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