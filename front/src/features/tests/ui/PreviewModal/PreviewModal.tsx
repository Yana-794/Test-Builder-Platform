import type { FC } from "react";
import { useState } from "react";
import styles from "./PreviewModal.module.css";
import type { EditorDraft } from "../../model/editorTypes";

type PreviewModalProps = {
  draft: EditorDraft;
  onClose: () => void;
};

export const PreviewModal: FC<PreviewModalProps> = ({ draft, onClose }) => {
  const [idx, setIdx] = useState(0);
  const q = draft.questions[idx] ?? null;
  const total = draft.questions.length;

  const hasIncomplete = draft.questions.some(
    (q) =>
      !q.prompt.trim() ||
      (q.type === "single_choice" &&
        (q.options.length < 2 || !q.options.some((o) => o.isCorrect))) ||
      (q.type === "text" && !(q.correctTextAnswer ?? "").trim()),
  );

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <header className={styles.header}>
          <div>
            <h2 className={styles.title}>
              {draft.meta.title || "Без названия"}
            </h2>
            {draft.meta.description && (
              <p className={styles.desc}>{draft.meta.description}</p>
            )}
          </div>
          <button className={styles.closeButton} onClick={onClose}>
            ×
          </button>
        </header>

        {hasIncomplete && (
          <div className={styles.warning}>
            Предварительный просмотр черновика. Некоторые вопросы не завершены.
          </div>
        )}

        {q ? (
          <div className={styles.body}>
            <div className={styles.progress}>
              Вопрос {idx + 1} из {total}
            </div>
            <h3 className={styles.prompt}>{q.prompt || "(без текста)"}</h3>
            {q.type === "single_choice" ? (
              <ul className={styles.options}>
                {q.options.map((o, i) => (
                  <li key={o.clientKey} className={styles.option}>
                    <span className={styles.radio}></span>
                    {String.fromCharCode(65 + i)}. {o.text || "(пусто)"}
                  </li>
                ))}
              </ul>
            ) : (
              <textarea
                disabled
                rows={3}
                placeholder="Ответ ученика..."
                className={styles.input}
              />
            )}
            <div className={styles.nav}>
              <button
                onClick={() => setIdx((i) => Math.max(0, i - 1))}
                disabled={idx === 0}
              >
                ← Назад
              </button>
              <button
                onClick={() => setIdx((i) => Math.min(total - 1, i + 1))}
                disabled={idx === total - 1}
              >
                Далее →
              </button>
            </div>
          </div>
        ) : (
          <p className={styles.empty}>Нет вопросов для предпросмотра</p>
        )}
      </div>
    </div>
  );
};
