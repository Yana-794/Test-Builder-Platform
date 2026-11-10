import type { FC } from "react";
import { useNavigate, useParams } from "react-router";
import { useAttemptResult } from "../../features/public-attempt/model/useAttemptResult";
import { ApiError } from "../../shared/api/http";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";
import mascotHappy from "../../assets/happy.png";
import mascotSad from "../../assets/sad.png";
import styles from "./PublicResultPage.module.css";

const GOOD_THRESHOLD = 60;

export const PublicResultPage: FC = () => {
  const { token = "", attemptId = "" } = useParams<{ token: string; attemptId: string }>();
  const navigate = useNavigate();
  const { data, isLoading, isError, error, refetch } = useAttemptResult(attemptId);

  if (isLoading) return <div className={styles.center}>Загрузка результата...</div>;

  if (isError || !data) {
    const notFound = error instanceof ApiError && error.status === 404;
    const notCompleted = error instanceof ApiError && error.status === 409;
    return (
      <div className={styles.card}>
        <ErrorState
          message={
            notFound ? "Результат не найден"
            : notCompleted ? "Тест ещё не завершён"
            : "Не удалось загрузить результат"
          }
          retryLabel={notFound ? "К тесту" : notCompleted ? "Продолжить" : "Повторить"}
          onRetry={
            notFound ? () => navigate(`/test/${token}`)
            : notCompleted ? () => navigate(`/test/${token}/attempt/${attemptId}`)
            : () => refetch()
          }
        />
      </div>
    );
  }

  const r = data.attempt;
  const isGood = r.percentage >= GOOD_THRESHOLD;
  const mascot = isGood ? mascotHappy : mascotSad;
  const minutes = Math.floor(r.durationSeconds / 60);
  const seconds = r.durationSeconds % 60;

  const message = isGood
    ? "Отличная работа! Вы справились с тестом."
    : "Спасибо за прохождение! Можно попробовать ещё раз.";

  return (
    <div className={styles.card}>
      <div className={styles.mascotWrap}>
        <img src={mascot} alt={isGood ? "Пуша радуется" : "Пуша грустит"} className={styles.mascot} />
      </div>

      <h1 className={styles.title}>Тест завершён</h1>
      <p className={styles.message}>{message}</p>

      <div className={styles.scoreBlock}>
        <div className={styles.percent}>{r.percentage}%</div>
        <div className={styles.scoreLine}>
          {r.score} из {r.maxScore} баллов
        </div>
      </div>

      <div className={styles.progressBar}>
        <div
          className={[styles.progressFill, isGood ? styles.fillGood : styles.fillSad].join(" ").trim()}
          style={{ width: `${Math.min(100, r.percentage)}%` }}
        />
      </div>

      <div className={styles.meta}>
        <div className={styles.metaRow}>
          <span className={styles.metaLabel}>Время прохождения</span>
          <span className={styles.metaValue}>
            {minutes > 0 ? `${minutes} мин ${seconds} сек` : `${seconds} сек`}
          </span>
        </div>
        <div className={styles.metaRow}>
          <span className={styles.metaLabel}>Завершено</span>
          <span className={styles.metaValue}>
            {new Date(r.completedAt).toLocaleString("ru-RU")}
          </span>
        </div>
      </div>

      <p className={styles.note}>
        Правильные ответы не раскрываются. Спасибо за прохождение!
      </p>

      <div className={styles.actions}>
        <button
          className={styles.secondary}
          onClick={() => navigate(`/test/${token}`, { replace: true })}
        >
          Пройти снова
        </button>
      </div>
    </div>
  );
};