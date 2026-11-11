import type { FC } from "react";
import { Hash, ListChecks, Type } from "lucide-react";
import type { QuestionType } from "../../../../entities/test/model/types";
import { OptionList } from "../OptionList/OptionList";
import styles from "./QuestionEditor.module.css";
import type { EditorOption, EditorQuestion } from "../../model/editorTypes";

type QuestionEditorProps = {
  question: EditorQuestion;
  index: number;
  onChange: (key: string, patch: Partial<EditorQuestion>) => void;
  onChangeType: (key: string, type: QuestionType) => void;
  onAddOption: (key: string) => void;
  onRemoveOption: (key: string, oKey: string) => void;
  onPatchOption: (
    key: string,
    oKey: string,
    patch: Partial<EditorOption>,
  ) => void;
  onSetCorrectOption: (key: string, oKey: string) => void;
  disabled?: boolean;
};

const TYPE_LABEL: Record<QuestionType, string> = {
  single_choice: "С вариантами",
  text: "Текстовый",
};

export const QuestionEditor: FC<QuestionEditorProps> = ({
  question,
  index,
  onChange,
  onChangeType,
  onAddOption,
  onRemoveOption,
  onPatchOption,
  onSetCorrectOption,
  disabled,
}) => {
  const promptLen = question.prompt.length;
  const promptNear = promptLen >= 1800;
  const textLen = (question.correctTextAnswer ?? "").length;
  const textNear = textLen >= 900;

  return (
    <div className={styles.editor}>
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <span className={styles.headerMark} aria-hidden="true" />
          <span className={styles.number}>Вопрос {index + 1}</span>
        </div>
        <span
          className={`${styles.typeChip} ${
            styles[`typeChip_${question.type}`] ?? ""
          }`}
        >
          {TYPE_LABEL[question.type] ?? question.type}
        </span>
      </div>

      <label className={styles.field}>
        <span className={styles.fieldLabel}>
          <Hash size={13} strokeWidth={2.4} aria-hidden="true" />
          Формулировка
        </span>
        <textarea
          className={styles.textarea}
          value={question.prompt}
          onChange={(e) =>
            onChange(question.clientKey, { prompt: e.target.value })
          }
          placeholder="Введите текст вопроса..."
          maxLength={2000}
          rows={3}
          disabled={disabled}
        />
        <span
          className={`${styles.counter} ${promptNear ? styles.counterNear : ""}`}
        >
          {promptLen} / 2000
        </span>
      </label>

      <label className={styles.field}>
        <span className={styles.fieldLabel}>
          <ListChecks size={13} strokeWidth={2.4} aria-hidden="true" />
          Тип вопроса
        </span>
        <div className={styles.selectWrap}>
          <select
            className={styles.select}
            value={question.type}
            onChange={(e) =>
              onChangeType(question.clientKey, e.target.value as QuestionType)
            }
            disabled={disabled}
          >
            <option value="single_choice">Один правильный вариант</option>
            <option value="text">Текстовый ответ</option>
          </select>
        </div>
      </label>

      {question.type === "single_choice" ? (
        <OptionList
          options={question.options}
          onAdd={() => onAddOption(question.clientKey)}
          onRemove={(oKey) => onRemoveOption(question.clientKey, oKey)}
          onPatch={(oKey, patch) =>
            onPatchOption(question.clientKey, oKey, patch)
          }
          onSetCorrect={(oKey) => onSetCorrectOption(question.clientKey, oKey)}
          disabled={disabled}
        />
      ) : (
        <label className={styles.field}>
          <span className={styles.fieldLabel}>
            <Type size={13} strokeWidth={2.4} aria-hidden="true" />
            Правильный текстовый ответ
          </span>
          <textarea
            className={styles.textarea}
            value={question.correctTextAnswer ?? ""}
            onChange={(e) =>
              onChange(question.clientKey, {
                correctTextAnswer: e.target.value,
              })
            }
            placeholder="Эталонный ответ (для проверки)"
            maxLength={1000}
            rows={2}
            disabled={disabled}
          />
          <span
            className={`${styles.counter} ${textNear ? styles.counterNear : ""}`}
          >
            {textLen} / 1000
          </span>
        </label>
      )}
    </div>
  );
};