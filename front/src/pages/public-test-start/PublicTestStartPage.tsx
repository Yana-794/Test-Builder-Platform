import type { FC } from "react";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";
import { usePublicTest } from "../../features/public-attempt/model/usePublicTest";
import { useCreateAttempt } from "../../features/public-attempt/model/useCreateAttempt";
import { ApiError } from "../../shared/api/http";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";
import styles from "./PublicTestStartPage.module.css";
import mascotWelcome from "../../assets/welcome.png";

export const PublicTestStartPage: FC = () => {
  const { token = "" } = useParams<{ token: string }>();
  const navigate = useNavigate();

  const { data, isLoading, isError, error, refetch } = usePublicTest(token);
  const create = useCreateAttempt(token);

  const [name, setName] = useState("");
  const [nameError, setNameError] = useState<string | null>(null);

  const handleStart = async () => {
    const trimmed = name.trim();
    if (!trimmed) {
      setNameError("Введите имя");
      return;
    }
    if (trimmed.length > 120) {
      setNameError("Слишком длинное имя (макс. 120)");
      return;
    }
    setNameError(null);
    try {
      const resp = await create.mutateAsync(trimmed);
      navigate(`/test/${token}/attempt/${resp.attempt.id}`, { replace: true });
    } catch {
     
    }
  };

  if (isLoading) return <div className={styles.center}>Загрузка...</div>;

  if (isError || !data) {
    const notFound = error instanceof ApiError && error.status === 404;
    return (
      <div className={styles.card}>
        <ErrorState
          message={notFound ? "Тест не найден или недоступен" : "Не удалось загрузить тест"}
          retryLabel={notFound ? undefined : "Повторить"}
          onRetry={notFound ? undefined : () => refetch()}
        />
      </div>
    );
  }

  const test = data.test;
  const createError =
    create.error instanceof ApiError
      ? create.error.status === 404
        ? "Тест больше недоступен"
        : "Не удалось начать тест. Попробуйте ещё раз."
      : null;

  return (
    <div className={styles.card}>
      <h1 className={styles.title}>{test.title}</h1>
      {test.description && <p className={styles.desc}>{test.description}</p>}

      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={styles.statValue}>{test.questionCount}</span>
          <span className={styles.statLabel}>вопросов</span>
        </div>
      </div>

      <p className={styles.hint}>
        Регистрация не требуется. Укажите имя - оно будет видно преподавателю
        в результатах.
      </p>

      <label className={styles.label}>
        Ваше имя
        <input
          type="text"
          className={styles.input}
          value={name}
          maxLength={120}
          onChange={(e) => { setName(e.target.value); setNameError(null); }}
          placeholder="Например, Анна"
          disabled={create.isPending}
          autoFocus
          onKeyDown={(e) => e.key === "Enter" && handleStart()}
        />
        {nameError && <span className={styles.fieldError}>{nameError}</span>}
      </label>

      {createError && <div className={styles.errorBox}>{createError}</div>}

      <button
        type="button"
        className={styles.primaryButton}
        onClick={handleStart}
        disabled={create.isPending}
      >
        {create.isPending ? "Начинаем..." : "Начать тест"}
      </button>
      <div className={styles.mascotWrap}>
        <img
          src={mascotWelcome}
          alt=""
          className={styles.mascot}
          aria-hidden
        />
      </div>
    </div>
  );
};