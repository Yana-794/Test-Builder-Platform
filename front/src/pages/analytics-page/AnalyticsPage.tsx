import type { FC } from "react";
import { useMemo, useState } from "react";
import { Link } from "react-router";
import { ArrowRight, BarChart3 } from "lucide-react";
import { useRecentAttempts } from "../../features/analytics/model/useRecentAttempts";
import { ApiError } from "../../shared/api/http";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";
import styles from "./AnalyticsPage.module.css";
import { SummaryCards } from "../../features/analytics/ui/SummaryCards/SummaryCards";
import { DistributionChart } from "../../features/analytics/ui/DistributionChart/DistributionChart";
import { RecentAttemptsTable } from "../../features/analytics/ui/RecentAttemptsTable/RecentAttemptsTable";
import { TestsFilterBar } from "../../features/analytics/ui/TestsFilterBar/TestsFilterBar";

function attemptsWord(n: number): string {
  const m10 = n % 10;
  const m100 = n % 100;
  if (m100 >= 11 && m100 <= 14) return "попыток";
  if (m10 === 1) return "попытка";
  if (m10 >= 2 && m10 <= 4) return "попытки";
  return "попыток";
}

export const AnalyticsPage: FC = () => {
  const { data, isLoading, isError, error, refetch } = useRecentAttempts(50);
  const [testId, setTestId] = useState<string | null>(null);

  const allAttempts = data?.attempts ?? [];

  const testsWithCounts = useMemo(() => {
    const map = new Map<string, { id: string; title: string; count: number }>();
    for (const a of allAttempts) {
      const existing = map.get(a.testId);
      if (existing) existing.count++;
      else map.set(a.testId, { id: a.testId, title: a.testTitle, count: 1 });
    }
    return Array.from(map.values()).sort((a, b) =>
      a.title.localeCompare(b.title, "ru"),
    );
  }, [allAttempts]);

  const attempts = useMemo(
    () =>
      testId ? allAttempts.filter((a) => a.testId === testId) : allAttempts,
    [allAttempts, testId],
  );

  if (isLoading) {
    return (
      <div className={styles.page}>
        <div className={styles.loading}>
          <span className={styles.spinner} aria-hidden="true" />
          <span>Загрузка аналитики...</span>
        </div>
      </div>
    );
  }

  if (isError) {
    const unauth = error instanceof ApiError && error.status === 401;
    return (
      <div className={styles.page}>
        <ErrorState
          message={
            unauth
              ? "Сессия истекла. Войдите заново."
              : "Не удалось загрузить аналитику"
          }
          retryLabel={unauth ? undefined : "Повторить"}
          onRetry={unauth ? undefined : () => refetch()}
        />
      </div>
    );
  }

  const selectedTest = testId
    ? testsWithCounts.find((t) => t.id === testId) ?? null
    : null;

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <span className={styles.heroIcon} aria-hidden="true">
          <BarChart3 size={22} strokeWidth={2.2} />
        </span>
        <div className={styles.headerContent}>
          <h1 className={styles.title}>Результаты</h1>
          <p className={styles.subtitle}>
            {selectedTest ? (
              <>
                Тест «
                <span className={styles.subtitleStrong}>
                  {selectedTest.title}
                </span>
                » · {attempts.length} {attemptsWord(attempts.length)}
              </>
            ) : (
              <>
                Все тесты · {attempts.length}{" "}
                {attemptsWord(attempts.length)} прохождения
              </>
            )}
          </p>
        </div>
      </header>

      {allAttempts.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyIcon} aria-hidden="true">
            <BarChart3 size={24} strokeWidth={2} />
          </span>
          <p className={styles.emptyTitle}>Попыток пока нет</p>
          <p className={styles.emptyText}>
            Опубликуйте тест и поделитесь ссылкой с учениками — здесь появятся
            результаты.
          </p>
          <Link to="/app/tests" className={styles.emptyLink}>
            К моим тестам
            <ArrowRight size={14} strokeWidth={2.4} aria-hidden="true" />
          </Link>
        </div>
      ) : (
        <>
          {testsWithCounts.length > 1 && (
            <TestsFilterBar
              tests={testsWithCounts}
              totalCount={allAttempts.length}
              selectedId={testId}
              onSelect={setTestId}
            />
          )}

          <SummaryCards attempts={attempts} />

          <section
            className={styles.section}
            aria-labelledby="distribution-heading"
          >
            <header className={styles.sectionHeader}>
              <div className={styles.sectionTitleWrap}>
                <span className={styles.sectionMark} aria-hidden="true" />
                <h2
                  className={styles.sectionTitle}
                  id="distribution-heading"
                >
                  Распределение результатов
                </h2>
              </div>
            </header>
            <DistributionChart attempts={attempts} />
          </section>

          <section
            className={styles.section}
            aria-labelledby="recent-attempts-heading"
          >
            <header className={styles.sectionHeader}>
              <div className={styles.sectionTitleWrap}>
                <span className={styles.sectionMark} aria-hidden="true" />
                <h2
                  className={styles.sectionTitle}
                  id="recent-attempts-heading"
                >
                  Последние попытки
                </h2>
              </div>
              <span className={styles.countChip}>{attempts.length}</span>
            </header>
            <RecentAttemptsTable attempts={attempts} />
          </section>
        </>
      )}
    </div>
  );
};