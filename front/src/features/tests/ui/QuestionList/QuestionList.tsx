import type { FC } from "react";
import { useState } from "react";
import type { QuestionType } from "../../../../entities/test/model/types";
import styles from "./QuestionList.module.css";
import type { EditorQuestion } from "../../model/editorTypes";

type QuestionListProps = {
  questions: EditorQuestion[];
  selectedKey: string | null;
  onSelect: (key: string) => void;
  onAdd: (type: QuestionType) => void;
  onMove: (key: string, direction: -1 | 1) => void;
  disabled?: boolean;
};

export const QuestionList: FC<QuestionListProps> = ({
  questions,
  selectedKey,
  onSelect,
  onAdd,
  onMove,
  disabled,
}) => {
  const [dragIdx, setDragIdx] = useState<number | null>(null);
  const [overIdx, setOverIdx] = useState<number | null>(null);

  const onDragStart = (e: React.DragEvent, idx: number) => {
    setDragIdx(idx);
    e.dataTransfer.effectAllowed = "move";
  };
  const onDragOver = (e: React.DragEvent, idx: number) => {
    e.preventDefault();
    if (overIdx !== idx) setOverIdx(idx);
  };
  const onDrop = (e: React.DragEvent, idx: number) => {
    e.preventDefault();
    if (dragIdx === null || dragIdx === idx) {
      setDragIdx(null);
      setOverIdx(null);
      return;
    }
    const key = questions[dragIdx].clientKey;
    const steps = Math.abs(idx - dragIdx);
    const dir: -1 | 1 = dragIdx < idx ? 1 : -1;
    for (let s = 0; s < steps; s++) onMove(key, dir);
    setDragIdx(null);
    setOverIdx(null);
  };

  return (
    <aside className={styles.list}>
      <div className={styles.header}>
        <span className={styles.title}>Вопросы</span>
        <span className={styles.count}>{questions.length}</span>
      </div>

      <ul className={styles.items}>
        {questions.map((q, idx) => (
          <li
            key={q.clientKey}
            className={[
              styles.item,
              q.clientKey === selectedKey ? styles.active : "",
              overIdx === idx ? styles.dragOver : "",
            ]
              .join(" ")
              .trim()}
            onClick={() => onSelect(q.clientKey)}
            draggable={!disabled}
            onDragStart={(e) => onDragStart(e, idx)}
            onDragOver={(e) => onDragOver(e, idx)}
            onDrop={(e) => onDrop(e, idx)}
            onDragEnd={() => {
              setDragIdx(null);
              setOverIdx(null);
            }}
          >
            <span className={styles.number}>{idx + 1}</span>
            <div className={styles.content}>
              <span className={styles.prompt}>
                {q.prompt.trim() || "(без текста)"}
              </span>
              <span className={styles.type}>
                {q.type === "single_choice" ? "Один ответ" : "Текст"}
              </span>
            </div>
            <div className={styles.controls}>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  onMove(q.clientKey, -1);
                }}
                disabled={disabled || idx === 0}
                title="Выше"
              >
                ↑
              </button>
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  onMove(q.clientKey, 1);
                }}
                disabled={disabled || idx === questions.length - 1}
                title="Ниже"
              >
                ↓
              </button>
            </div>
          </li>
        ))}
      </ul>

      <div className={styles.actions}>
        <button
          className={styles.addButton}
          onClick={() => onAdd("single_choice")}
          disabled={disabled || questions.length >= 100}
        >
          + Вопрос с вариантами
        </button>
        <button
          className={styles.addButton}
          onClick={() => onAdd("text")}
          disabled={disabled || questions.length >= 100}
        >
          + Текстовый вопрос
        </button>
      </div>
    </aside>
  );
};
