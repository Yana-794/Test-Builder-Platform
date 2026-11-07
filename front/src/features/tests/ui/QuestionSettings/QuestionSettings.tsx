import type { FC } from "react";
import styles from "./QuestionSettings.module.css";
import type { EditorQuestion } from "../../model/editorTypes";

type QuestionSettingsProps = {
  question: EditorQuestion;
  index: number;
  total: number;
  onChange: (key: string, patch: Partial<EditorQuestion>) => void;
  onRemove: (key: string) => void;
  onMove: (key: string, direction: -1 | 1) => void;
  disabled?: boolean;
};

export const QuestionSettings: FC<QuestionSettingsProps> = ({
  question,
  index,
  total,
  onChange,
  onRemove,
  onMove,
  disabled,
}) => {
  return (
    <div className={styles.panel}>
      <h2 className={styles.title}>Настройки вопроса</h2>

      <label className={styles.label}>
        Баллы
        <input
          type="number"
          className={styles.input}
          value={question.points}
          step={0.25}
          min={0.01}
          onChange={(e) => {
            const v = parseFloat(e.target.value);
            if (!Number.isNaN(v)) onChange(question.clientKey, { points: v });
          }}
          disabled={disabled}
        />
      </label>

      <div className={styles.block}>
        <span className={styles.blockLabel}>Порядок</span>
        <div className={styles.orderButtons}>
          <button
            className={styles.orderButton}
            onClick={() => onMove(question.clientKey, -1)}
            disabled={disabled || index === 0}
          >
            ↑ Выше
          </button>
          <button
            className={styles.orderButton}
            onClick={() => onMove(question.clientKey, 1)}
            disabled={disabled || index === total - 1}
          >
            ↓ Ниже
          </button>
        </div>
      </div>

      <button
        className={styles.deleteButton}
        onClick={() => onRemove(question.clientKey)}
        disabled={disabled}
      >
        Удалить вопрос
      </button>
    </div>
  );
};
