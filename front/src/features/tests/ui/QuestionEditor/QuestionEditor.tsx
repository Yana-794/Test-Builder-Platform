import type { FC } from "react";
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
  return (
    <div className={styles.editor}>
      <div className={styles.header}>
        <span className={styles.number}>Вопрос {index + 1}</span>
      </div>

      <label className={styles.label}>
        Формулировка
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
        <span className={styles.counter}>{question.prompt.length} / 2000</span>
      </label>

      <label className={styles.label}>
        Тип вопроса
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
        <label className={styles.label}>
          Правильный текстовый ответ
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
          <span className={styles.counter}>
            {(question.correctTextAnswer ?? "").length} / 1000
          </span>
        </label>
      )}
    </div>
  );
};
