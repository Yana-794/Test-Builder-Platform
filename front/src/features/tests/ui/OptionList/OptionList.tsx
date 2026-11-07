import type { FC } from "react";
import styles from "./OptionList.module.css";
import type { EditorOption } from "../../model/editorTypes";

type OptionListProps = {
  options: EditorOption[];
  onAdd: () => void;
  onRemove: (oKey: string) => void;
  onPatch: (oKey: string, patch: Partial<EditorOption>) => void;
  onSetCorrect: (oKey: string) => void;
  disabled?: boolean;
};

export const OptionList: FC<OptionListProps> = ({
  options,
  onAdd,
  onRemove,
  onPatch,
  onSetCorrect,
  disabled,
}) => {
  const canRemove = options.length > 2;

  return (
    <div className={styles.list}>
      <div className={styles.header}>
        <span>Варианты ответа</span>
        <span className={styles.hint}>Отметьте один правильный</span>
      </div>
      <ul className={styles.items}>
        {options.map((o, idx) => (
          <li key={o.clientKey} className={styles.item}>
            <input
              type="radio"
              name="correct-option"
              checked={o.isCorrect}
              onChange={() => onSetCorrect(o.clientKey)}
              disabled={disabled}
              className={styles.radio}
            />
            <span className={styles.letter}>
              {String.fromCharCode(65 + idx)}
            </span>
            <input
              type="text"
              className={styles.input}
              value={o.text}
              onChange={(e) => onPatch(o.clientKey, { text: e.target.value })}
              placeholder={`Вариант ${idx + 1}`}
              maxLength={500}
              disabled={disabled}
            />
            <button
              type="button"
              className={styles.removeButton}
              onClick={() => onRemove(o.clientKey)}
              disabled={disabled || !canRemove}
              title={canRemove ? "Удалить" : "Минимум 2 варианта"}
            >
              ×
            </button>
          </li>
        ))}
      </ul>
      <button
        type="button"
        className={styles.addButton}
        onClick={onAdd}
        disabled={disabled || options.length >= 10}
      >
        + Добавить вариант
      </button>
    </div>
  );
};
