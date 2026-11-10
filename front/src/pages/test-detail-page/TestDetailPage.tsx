import type { FC } from "react";
import { useEffect } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { QRCodeSVG } from "qrcode.react";
import { getTest } from "../../features/tests/api/getTest";
import { testsQueryKeys } from "../../features/tests/model/queryKeys";
import { ApiError } from "../../shared/api/http";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";
import styles from "./TestDetailPage.module.css";
import { usePublication } from "../../features/tests/model/hooks/usePublication";

const STATUS_LABEL: Record<string, string> = {
  draft: "Черновик",
  published: "Опубликован",
  archived: "Архив",
};

export const TestDetailPage: FC = () => {
  const { testId = "" } = useParams<{ testId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: testsQueryKeys.detail(testId),
    queryFn: ({ signal }) => getTest(testId, signal),
    retry: false,
    staleTime: 30_000,
  });

  const isPublished = data?.test.status === "published";
  const publicationQuery = usePublication(testId, isPublished);

  useEffect(() => {
    if (error instanceof ApiError && error.status === 401) {
      navigate("/login", { replace: true });
    }
  }, [error, navigate]);

  if (isLoading) return <div className={styles.loading}>Загрузка...</div>;

  if (error instanceof ApiError && error.status === 404) {
    return (
      <ErrorState
        message="Тест не найден или недоступен"
        retryLabel="К списку тестов"
        onRetry={() => navigate("/app/tests")}
      />
    );
  }

  if (isError || !data) {
    return <ErrorState message="Не удалось загрузить тест" onRetry={() => refetch()} />;
  }

  const test = data.test;
  const isDraft = test.status === "draft";
  const totalPoints = test.questions.reduce((s, q) => s + q.points, 0);
  const pub = publicationQuery.data?.publication ?? null;

  const handleCopy = async () => {
    if (!pub) return;
    try {
      await navigator.clipboard.writeText(pub.publicUrl);
    } catch {
      // игнорируем — пользователь скопирует вручную
    }
  };

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Link to="/app/tests" className={styles.back}>← К списку</Link>
        <div className={styles.headerActions}>
          {isDraft && (
            <Link to={`/app/tests/${test.id}/edit`} className={styles.primaryButton}>
              Редактировать
            </Link>
          )}
        </div>
      </header>

      <div className={styles.titleRow}>
        <h1 className={styles.title}>{test.title || "(без названия)"}</h1>
        <span className={`${styles.badge} ${styles[`badge_${test.status}`]}`}>
          {STATUS_LABEL[test.status]}
        </span>
      </div>

      {test.description && <p className={styles.desc}>{test.description}</p>}

      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={styles.statValue}>{test.questions.length}</span>
          <span className={styles.statLabel}>вопросов</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statValue}>{totalPoints}</span>
          <span className={styles.statLabel}>баллов</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statValue}>v{test.revision}</span>
          <span className={styles.statLabel}>версия</span>
        </div>
        {test.subject && (
          <div className={styles.stat}>
            <span className={styles.statValue}>{test.subject}</span>
            <span className={styles.statLabel}>предмет</span>
          </div>
        )}
      </div>

      
      {isPublished && (
        <section className={styles.publicationBlock}>
          <h2 className={styles.sectionTitle}>Публикация</h2>

          {publicationQuery.isLoading && <p className={styles.muted}>Загрузка ссылки...</p>}

          {publicationQuery.isError && (
            <p className={styles.muted}>
              Не удалось загрузить ссылку.{" "}
              <button onClick={() => publicationQuery.refetch()} className={styles.link}>
                Повторить
              </button>
            </p>
          )}

          {pub && (
            <div className={styles.publicationContent}>
              <div className={styles.qrWrap}>
                <QRCodeSVG value={pub.publicUrl} size={160} includeMargin />
              </div>
              <div className={styles.pubMeta}>
                <div className={styles.pubRow}>
                  <span className={styles.muted}>Публичная ссылка</span>
                  <span className={styles.pubVersion}>
                    версия {pub.version} · {new Date(pub.publishedAt).toLocaleString("ru-RU")}
                  </span>
                </div>
                <div className={styles.linkRow}>
                  <input
                    className={styles.linkInput}
                    value={pub.publicUrl}
                    readOnly
                    onFocus={(e) => e.currentTarget.select()}
                  />
                  <button className={styles.copyButton} onClick={handleCopy}>
                    Копировать
                  </button>
                  <button
                    className={styles.openButton}
                    onClick={() =>
                      window.open(pub.publicUrl, "_blank", "noopener,noreferrer")
                    }
                  >
                    Открыть
                  </button>
                </div>
              </div>
            </div>
          )}
        </section>
      )}

      <section className={styles.questions}>
        <h2 className={styles.sectionTitle}>Вопросы</h2>
        {test.questions.length === 0 ? (
          <p className={styles.noQuestions}>
            Вопросов пока нет.{" "}
            {isDraft && (
              <Link to={`/app/tests/${test.id}/edit`} className={styles.link}>
                Добавить
              </Link>
            )}
          </p>
        ) : (
          <ol className={styles.questionsList}>
            {test.questions.map((q) => (
              <li key={q.id} className={styles.question}>
                <div className={styles.questionHeader}>
                  <span className={styles.questionType}>
                    {q.type === "single_choice" ? "Один ответ" : "Текстовый ответ"}
                  </span>
                  <span className={styles.questionPoints}>{q.points} б.</span>
                </div>
                <p className={styles.questionPrompt}>{q.prompt || "(без текста)"}</p>
                {q.type === "single_choice" && q.options && (
                  <ul className={styles.options}>
                    {q.options.map((o) => (
                      <li
                        key={o.id}
                        className={[styles.option, o.isCorrect ? styles.optionCorrect : ""].join(" ").trim()}
                      >
                        {o.isCorrect && <span className={styles.checkMark}>✓</span>}
                        {o.text || "(пусто)"}
                      </li>
                    ))}
                  </ul>
                )}
                {q.type === "text" && q.correctTextAnswer && (
                  <p className={styles.correctAnswer}>
                    Правильный ответ: <b>{q.correctTextAnswer}</b>
                  </p>
                )}
              </li>
            ))}
          </ol>
        )}
      </section>
    </div>
  );
};