import type { FC } from "react";
import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { QRCodeSVG } from "qrcode.react";
import {
  ArrowLeft,
  Award,
  BookOpen,
  Check,
  CheckCircle2,
  Clipboard,
  ClipboardCheck,
  ExternalLink,
  GitBranch,
  ListChecks,
  Pencil,
  
} from "lucide-react";
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

  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (error instanceof ApiError && error.status === 401) {
      navigate("/login", { replace: true });
    }
  }, [error, navigate]);

  if (isLoading) {
    return (
      <div className={styles.page}>
        <div className={styles.loading}>
          <span className={styles.spinner} aria-hidden="true" />
          Загрузка теста...
        </div>
      </div>
    );
  }

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
    return (
      <ErrorState
        message="Не удалось загрузить тест"
        onRetry={() => refetch()}
      />
    );
  }

  const test = data.test;
  const isDraft = test.status === "draft";
  const totalPoints = test.questions.reduce((s, q) => s + q.points, 0);
  const pub = publicationQuery.data?.publication ?? null;

  const handleCopy = async () => {
    if (!pub) return;
    try {
      await navigator.clipboard.writeText(pub.publicUrl);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      /* ignore */
    }
  };

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Link to="/app/tests" className={styles.back}>
          <ArrowLeft size={16} strokeWidth={2.2} aria-hidden="true" />
          <span>К списку</span>
        </Link>

        <div className={styles.headerActions}>
          {isDraft && (
            <Link
              to={`/app/tests/${test.id}/edit`}
              className={styles.primaryButton}
            >
              <Pencil size={15} strokeWidth={2.2} />
              Редактировать
            </Link>
          )}
        </div>
      </header>

      <div className={styles.titleBlock}>
        <div className={styles.titleRow}>
          <h1 className={styles.title}>{test.title || "(без названия)"}</h1>
          <span
            className={`${styles.badge} ${styles[`badge_${test.status}`] ?? ""}`}
          >
            {STATUS_LABEL[test.status] ?? test.status}
          </span>
        </div>
        {test.description && <p className={styles.desc}>{test.description}</p>}
      </div>

      <div className={styles.stats}>
        <div className={styles.stat}>
          <span className={`${styles.statIcon} ${styles.statIcon_primary}`} aria-hidden="true">
            <ListChecks size={18} strokeWidth={2.2} />
          </span>
          <div className={styles.statBody}>
            <span className={styles.statValue}>{test.questions.length}</span>
            <span className={styles.statLabel}>вопросов</span>
          </div>
        </div>

        <div className={styles.stat}>
          <span className={`${styles.statIcon} ${styles.statIcon_primary}`} aria-hidden="true">
            <Award size={18} strokeWidth={2.2} />
          </span>
          <div className={styles.statBody}>
            <span className={styles.statValue}>{totalPoints}</span>
            <span className={styles.statLabel}>баллов</span>
          </div>
        </div>

        <div className={styles.stat}>
          <span className={`${styles.statIcon} ${styles.statIcon_neutral}`} aria-hidden="true">
            <GitBranch size={18} strokeWidth={2.2} />
          </span>
          <div className={styles.statBody}>
            <span className={styles.statValue}>v{test.revision}</span>
            <span className={styles.statLabel}>версия</span>
          </div>
        </div>

        {test.subject && (
          <div className={styles.stat}>
            <span className={`${styles.statIcon} ${styles.statIcon_neutral}`} aria-hidden="true">
              <BookOpen size={18} strokeWidth={2.2} />
            </span>
            <div className={styles.statBody}>
              <span className={styles.statValue}>{test.subject}</span>
              <span className={styles.statLabel}>предмет</span>
            </div>
          </div>
        )}
      </div>

      {isPublished && (
        <section className={styles.publicationBlock} aria-label="Публикация">

          {publicationQuery.isLoading && (
            <p className={styles.muted}>Загрузка ссылки...</p>
          )}

          {publicationQuery.isError && (
            <p className={styles.muted}>
              Не удалось загрузить ссылку.{" "}
              <button
                type="button"
                onClick={() => publicationQuery.refetch()}
                className={styles.link}
              >
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
                  <span className={styles.pubLabel}>Публичная ссылка</span>
                  <span className={styles.pubVersion}>
                    версия {pub.version} ·{" "}
                    {new Date(pub.publishedAt).toLocaleString("ru-RU")}
                  </span>
                </div>

                <div className={styles.linkRow}>
                  <input
                    className={styles.linkInput}
                    value={pub.publicUrl}
                    readOnly
                    onFocus={(e) => e.currentTarget.select()}
                    aria-label="Публичная ссылка"
                  />
                  <button
                    type="button"
                    className={`${styles.copyButton} ${copied ? styles.copyButtonActive : ""}`}
                    onClick={handleCopy}
                  >
                    {copied ? (
                      <ClipboardCheck size={15} strokeWidth={2.2} />
                    ) : (
                      <Clipboard size={15} strokeWidth={2.2} />
                    )}
                    <span>{copied ? "Скопировано" : "Копировать"}</span>
                  </button>
                  <button
                    type="button"
                    className={styles.openButton}
                    onClick={() =>
                      window.open(
                        pub.publicUrl,
                        "_blank",
                        "noopener,noreferrer",
                      )
                    }
                  >
                    <ExternalLink size={15} strokeWidth={2.2} />
                    <span>Открыть</span>
                  </button>
                </div>
              </div>
            </div>
          )}
        </section>
      )}

      <section className={styles.questions} aria-label="Вопросы">
        <header className={styles.sectionHeader}>
          <div className={styles.sectionTitleWrap}>
            <span className={styles.sectionMark} aria-hidden="true" />
            <h2 className={styles.sectionTitle}>Вопросы</h2>
          </div>
          {test.questions.length > 0 && (
            <span className={styles.questionsCount}>
              {test.questions.length}
            </span>
          )}
        </header>

        {test.questions.length === 0 ? (
          <div className={styles.emptyQuestions}>
            <span className={styles.emptyIcon} aria-hidden="true">
              <ListChecks size={22} strokeWidth={2} />
            </span>
            <p className={styles.emptyTitle}>Вопросов пока нет</p>
            {isDraft && (
              <Link
                to={`/app/tests/${test.id}/edit`}
                className={styles.linkInline}
              >
                Добавить вопрос
              </Link>
            )}
          </div>
        ) : (
          <ol className={styles.questionsList}>
            {test.questions.map((q, idx) => (
              <li key={q.id} className={styles.question}>
                <div className={styles.questionHeader}>
                  <span className={styles.questionNum}>{idx + 1}</span>
                  <span
                    className={`${styles.questionType} ${
                      styles[`questionType_${q.type}`] ?? ""
                    }`}
                  >
                    {q.type === "single_choice"
                      ? "Один ответ"
                      : "Текстовый ответ"}
                  </span>
                  <span className={styles.questionPoints}>{q.points} б.</span>
                </div>

                <p className={styles.questionPrompt}>
                  {q.prompt || "(без текста)"}
                </p>

                {q.type === "single_choice" && q.options && (
                  <ul className={styles.options}>
                    {q.options.map((o) => (
                      <li
                        key={o.id}
                        className={`${styles.option} ${
                          o.isCorrect ? styles.optionCorrect : ""
                        }`}
                      >
                        <span
                          className={styles.optionMark}
                          aria-hidden="true"
                        >
                          {o.isCorrect ? (
                            <Check size={12} strokeWidth={3} />
                          ) : null}
                        </span>
                        <span className={styles.optionText}>
                          {o.text || "(пусто)"}
                        </span>
                      </li>
                    ))}
                  </ul>
                )}

                {q.type === "text" && q.correctTextAnswer && (
                  <div className={styles.correctAnswer}>
                    <CheckCircle2
                      size={14}
                      strokeWidth={2.4}
                      aria-hidden="true"
                    />
                    <span>
                      Правильный ответ: <b>{q.correctTextAnswer}</b>
                    </span>
                  </div>
                )}
              </li>
            ))}
          </ol>
        )}
      </section>
    </div>
  );
};