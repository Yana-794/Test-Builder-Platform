import type { FC } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router";
import { useAttemptState } from "../../features/public-attempt/model/useAttemptState";
import { useSaveAnswers } from "../../features/public-attempt/model/useSaveAnswers";
import type { AnswerDTO, PublicQuestion } from "../../entities/test/model/types";
import { ApiError } from "../../shared/api/http";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";
import styles from "./PublicAttemptPage.module.css";
import { useSubmitAttempt } from "../../features/public-attempt/model/useSubmitAttempt";

const SAVE_DEBOUNCE_MS = 800;

export const PublicAttemptPage: FC = () => {
  const { token = "", attemptId = "" } = useParams<{ token: string; attemptId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, isError, error, refetch } = useAttemptState(attemptId);
  const save = useSaveAnswers(attemptId);
  const submit = useSubmitAttempt(attemptId);

  const [answers, setAnswers] = useState<Record<string, AnswerDTO>>({});
  const [currentIdx, setCurrentIdx] = useState(0);
  const [submitOpen, setSubmitOpen] = useState(false);
  const [saveError, setSaveError] = useState(false);

  const initialized = useRef(false);
  useEffect(() => {
    if (!data || initialized.current) return;
    const init: Record<string, AnswerDTO> = {};
    for (const a of data.answers) init[a.questionId] = a;
    setAnswers(init);
    initialized.current = true;
  }, [data]);

  // Если попытка уже завершена, то сразу на результат
  useEffect(() => {
    if (data?.attempt.status === "completed") {
      navigate(`/test/${token}/result/${attemptId}`, { replace: true });
    }
  }, [data?.attempt.status, navigate, token, attemptId]);

  const pendingRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const flush = useCallback(() => {
    return new Promise<void>((resolve, reject) => {
      const payload = { answers: Object.values(answers) };
      save.mutate(payload, {
        onSuccess: () => resolve(),
        onError: (err) => { setSaveError(true); reject(err); },
      });
    });
  }, [answers, save]);

  const scheduleSave = useCallback(() => {
    setSaveError(false);
    if (pendingRef.current) clearTimeout(pendingRef.current);
    pendingRef.current = setTimeout(() => {
      const payload = { answers: Object.values(answers) };
      save.mutate(payload, { onError: () => setSaveError(true) });
      pendingRef.current = null;
    }, SAVE_DEBOUNCE_MS);
  }, [answers, save]);

  useEffect(() => () => { if (pendingRef.current) clearTimeout(pendingRef.current); }, []);

  const questions = useMemo(() => data?.test.questions ?? [], [data]);
  const current = questions[currentIdx] ?? null;
  const total = questions.length;
  const answered = Object.values(answers).filter(
    (a) => a.selectedOptionId || (a.answerText && a.answerText.trim()),
  ).length;

  const updateSingle = (q: PublicQuestion, optionId: string) => {
    setAnswers((p) => ({ ...p, [q.id]: { questionId: q.id, selectedOptionId: optionId } }));
    scheduleSave();
  };

  const updateText = (q: PublicQuestion, text: string) => {
    setAnswers((p) => ({ ...p, [q.id]: { questionId: q.id, answerText: text } }));
    scheduleSave();
  };

  const handleSubmit = async () => {
    if (pendingRef.current) {
      clearTimeout(pendingRef.current);
      pendingRef.current = null;
    }
    try {
      await flush();
      await submit.mutateAsync();
      navigate(`/test/${token}/result/${attemptId}`, { replace: true });
    } catch {
    }
  };

  if (isLoading) return <div className={styles.center}>Загрузка...</div>;

  if (isError || !data) {
    const notFound = error instanceof ApiError && error.status === 404;
    return (
      <div className={styles.wrap}>
        <ErrorState
          message={notFound ? "Попытка не найдена" : "Не удалось загрузить попытку"}
          retryLabel={notFound ? "Открыть тест" : "Повторить"}
          onRetry={notFound ? () => navigate(`/test/${token}`) : () => refetch()}
        />
      </div>
    );
  }

  return (
    <div className={styles.wrap}>
      <header className={styles.header}>
        <h1 className={styles.title}>{data.test.title}</h1>
        <span className={styles.progress}>
          {currentIdx + 1} / {total}
        </span>
      </header>

      <div className={styles.progressBar}>
        <div
          className={styles.progressFill}
          style={{ width: `${((currentIdx + 1) / total) * 100}%` }}
        />
      </div>

      {current && (
        <main className={styles.question}>
          <span className={styles.qLabel}>Вопрос {currentIdx + 1}</span>
          <h2 className={styles.qPrompt}>{current.prompt || "(без текста)"}</h2>

          {current.type === "single_choice" && current.options && (
            <ul className={styles.options}>
              {current.options.map((o) => {
                const isSel = answers[current.id]?.selectedOptionId === o.id;
                return (
                  <li key={o.id}>
                    <button
                      type="button"
                      className={[styles.option, isSel ? styles.optionSelected : ""].join(" ").trim()}
                      onClick={() => updateSingle(current, o.id)}
                    >
                      <span className={styles.radio}>{isSel ? "●" : ""}</span>
                      {o.text || "(пусто)"}
                    </button>
                  </li>
                );
              })}
            </ul>
          )}

          {current.type === "text" && (
            <textarea
              className={styles.textarea}
              value={answers[current.id]?.answerText ?? ""}
              onChange={(e) => updateText(current, e.target.value)}
              placeholder="Ваш ответ..."
              rows={4}
            />
          )}
        </main>
      )}

      <footer className={styles.footer}>
        <button
          className={styles.secondary}
          onClick={() => { flush().catch(() => {}); setCurrentIdx((i) => Math.max(0, i - 1)); }}
          disabled={currentIdx === 0}
        >
          ← Назад
        </button>

        {currentIdx < total - 1 ? (
          <button
            className={styles.primary}
            onClick={() => { flush().catch(() => {}); setCurrentIdx((i) => i + 1); }}
          >
            Далее 
          </button>
        ) : (
          <button className={styles.submit} onClick={() => setSubmitOpen(true)}>
            Завершить тест
          </button>
        )}
      </footer>

      <div className={styles.statusLine}>
        {save.isPending && <span>Сохранение...</span>}
        {saveError && <span className={styles.statusError}>Ошибка сохранения</span>}
        {!save.isPending && !saveError && (
          <span>Отвечено: {answered} из {total}</span>
        )}
      </div>

      {submitOpen && (
        <div className={styles.overlay} onClick={() => setSubmitOpen(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 className={styles.modalTitle}>Завершить тест?</h2>
            <p className={styles.modalText}>
              Отвечено {answered} из {total} вопросов. После завершения изменить ответы будет нельзя.
            </p>
            <div className={styles.modalActions}>
              <button className={styles.secondary} onClick={() => setSubmitOpen(false)}>
                Остаться
              </button>
              <button
                className={styles.primary}
                onClick={handleSubmit}
                disabled={submit.isPending}
              >
                {submit.isPending ? "Отправка..." : "Завершить"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};