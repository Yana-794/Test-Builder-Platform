import type { FC } from "react";
import { useState } from "react";
import {
  ChevronDown,
  ChevronUp,
  GripVertical,
  ListPlus,
  Type,
} from "lucide-react";
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

const TYPE_LABEL: Record<string, string> = {
  single_choice: "Один ответ",
  text: "Текст",
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
    <aside className={styles.list} aria-label="Список вопросов">
      <div className={styles.header}>
        <div className={styles.titleWrap}>
          <span className={styles.titleMark} aria-hidden="true" />
          <span className={styles.title}>Вопросы</span>
        </div>
        <span
          className={styles.count}
          aria-label={`Всего: ${questions.length}`}
        >
          {questions.length}
        </span>
      </div>

      <ul className={styles.items} role="listbox" aria-label="Вопросы теста">
        {questions.length === 0 && (
          <li className={styles.empty} role="presentation">
            <span className={styles.emptyTitle}>Пока нет вопросов</span>
            <span className={styles.emptyText}>
              Добавьте первый вопрос ниже.
            </span>
          </li>
        )}

        {questions.map((q, idx) => {
          const isActive = q.clientKey === selectedKey;
          const isOver = overIdx === idx;
          return (
            <li
              key={q.clientKey}
              role="option"
              aria-selected={isActive}
              tabIndex={disabled ? -1 : 0}
              className={[
                styles.item,
                isActive ? styles.active : "",
                isOver ? styles.dragOver : "",
                disabled ? styles.itemDisabled : "",
              ]
                .join(" ")
                .trim()}
              onClick={() => onSelect(q.clientKey)}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  onSelect(q.clientKey);
                }
              }}
              draggable={!disabled}
              onDragStart={(e) => onDragStart(e, idx)}
              onDragOver={(e) => onDragOver(e, idx)}
              onDrop={(e) => onDrop(e, idx)}
              onDragEnd={() => {
                setDragIdx(null);
                setOverIdx(null);
              }}
            >
              <span className={styles.grip} aria-hidden="true">
                <GripVertical size={14} strokeWidth={2.2} />
              </span>

              <span className={styles.number}>{idx + 1}</span>

              <div className={styles.content}>
                <span className={styles.prompt}>
                  {q.prompt.trim() || "(без текста)"}
                </span>
                <span
                  className={`${styles.type} ${styles[`type_${q.type}`] ?? ""}`}
                >
                  {TYPE_LABEL[q.type] ?? q.type}
                </span>
              </div>

              <div className={styles.controls}>
                <button
                  type="button"
                  className={styles.controlBtn}
                  onClick={(e) => {
                    e.stopPropagation();
                    onMove(q.clientKey, -1);
                  }}
                  disabled={disabled || idx === 0}
                  title="Выше"
                  aria-label="Переместить выше"
                >
                  <ChevronUp size={14} strokeWidth={2.4} />
                </button>
                <button
                  type="button"
                  className={styles.controlBtn}
                  onClick={(e) => {
                    e.stopPropagation();
                    onMove(q.clientKey, 1);
                  }}
                  disabled={disabled || idx === questions.length - 1}
                  title="Ниже"
                  aria-label="Переместить ниже"
                >
                  <ChevronDown size={14} strokeWidth={2.4} />
                </button>
              </div>
            </li>
          );
        })}
      </ul>

      <div className={styles.actions}>
        <button
          type="button"
          className={styles.addButton}
          onClick={() => onAdd("single_choice")}
          disabled={disabled || questions.length >= 100}
        >
          <ListPlus size={15} strokeWidth={2.2} />
          <span>Вопрос с вариантами</span>
        </button>
        <button
          type="button"
          className={styles.addButton}
          onClick={() => onAdd("text")}
          disabled={disabled || questions.length >= 100}
        >
          <Type size={15} strokeWidth={2.2} />
          <span>Текстовый вопрос</span>
        </button>
      </div>
    </aside>
  );
};
