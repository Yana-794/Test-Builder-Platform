import type { FC } from "react";
import { useMemo, useState } from "react";
import { Link } from "react-router";
import {
  formatDuration,
  formatRelative,
} from "../../../../features/analytics/lib/format";
import styles from "./RecentAttemptsTable.module.css";
import type { AnalyticsRecentAttempt } from "../../../../entities/analytics";

type SortField =
  | "studentName"
  | "testTitle"
  | "status"
  | "percent"
  | "completedAt";
type SortDir = "asc" | "desc";
type StatusFilter = "all" | "completed" | "in_progress";
type ResultMode = "percent" | "points";

type Props = { attempts: AnalyticsRecentAttempt[] };

const STATUS_TABS: { v: StatusFilter; label: string }[] = [
  { v: "all", label: "Все" },
  { v: "completed", label: "Завершённые" },
  { v: "in_progress", label: "В процессе" },
];

const RESULT_MODES: { v: ResultMode; label: string }[] = [
  { v: "percent", label: "%" },
  { v: "points", label: "Баллы" },
];

export const RecentAttemptsTable: FC<Props> = ({ attempts }) => {
  const [sortField, setSortField] = useState<SortField>("completedAt");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [resultMode, setResultMode] = useState<ResultMode>("percent");

  const filtered = useMemo(() => {
    const list =
      statusFilter === "all"
        ? attempts
        : attempts.filter((a) => a.status === statusFilter);
    const dir = sortDir === "asc" ? 1 : -1;
    return [...list].sort((a, b) => {
      const va = sortValue(a, sortField);
      const vb = sortValue(b, sortField);
      if (va === vb) return 0;
      return va < vb ? -dir : dir;
    });
  }, [attempts, statusFilter, sortField, sortDir]);

  const toggleSort = (f: SortField) => {
    if (sortField === f) setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    else {
      setSortField(f);
      setSortDir("desc");
    }
  };

  const arrow = (f: SortField) =>
    sortField === f ? (sortDir === "asc" ? "↑" : "↓") : "";

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        <div className={styles.tabs} role="tablist">
          {STATUS_TABS.map((t) => (
            <button
              key={t.v}
              type="button"
              role="tab"
              aria-selected={statusFilter === t.v}
              className={`${styles.tab} ${statusFilter === t.v ? styles.tabActive : ""}`}
              onClick={() => setStatusFilter(t.v)}
            >
              {t.label}
            </button>
          ))}
        </div>

        <div
          className={styles.segmented}
          role="group"
          aria-label="Формат результата"
        >
          {RESULT_MODES.map((m) => (
            <button
              key={m.v}
              type="button"
              role="radio"
              aria-checked={resultMode === m.v}
              className={`${styles.segment} ${resultMode === m.v ? styles.segmentActive : ""}`}
              onClick={() => setResultMode(m.v)}
            >
              {m.label}
            </button>
          ))}
        </div>
      </div>

      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th
                onClick={() => toggleSort("studentName")}
                className={styles.sortable}
              >
                Ученик
                <span className={styles.arrow}>{arrow("studentName")}</span>
              </th>
              <th
                onClick={() => toggleSort("testTitle")}
                className={styles.sortable}
              >
                Тест
                <span className={styles.arrow}>{arrow("testTitle")}</span>
              </th>
              <th
                onClick={() => toggleSort("status")}
                className={styles.sortable}
              >
                Статус
                <span className={styles.arrow}>{arrow("status")}</span>
              </th>
              <th
                onClick={() => toggleSort("percent")}
                className={styles.sortable}
              >
                Результат
                <span className={styles.arrow}>{arrow("percent")}</span>
              </th>
              <th>Время</th>
              <th
                onClick={() => toggleSort("completedAt")}
                className={styles.sortable}
              >
                Завершено
                <span className={styles.arrow}>{arrow("completedAt")}</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((a) => {
              const isDone = a.status === "completed";
              return (
                <tr key={a.attemptId}>
                  <td className={styles.studentName}>{a.studentName}</td>
                  <td>
                    <Link to={`/app/tests/${a.testId}`} className={styles.link}>
                      {a.testTitle}
                    </Link>
                  </td>
                  <td>
                    <span
                      className={`${styles.status} ${isDone ? styles.statusDone : styles.statusProgress}`}
                    >
                      <span className={styles.statusDot} aria-hidden="true" />
                      {isDone ? "Завершён" : "В процессе"}
                    </span>
                  </td>
                  <td className={`${styles.num} ${percentClass(a.percent)}`}>
                    {renderResult(a, resultMode)}
                  </td>
                  <td className={`${styles.num} ${styles.muted}`}>
                    {formatDuration(a.durationSeconds)}
                  </td>
                  <td className={styles.muted}>
                    {isDone
                      ? formatRelative(a.completedAt)
                      : formatRelative(a.startedAt)}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>

        {filtered.length === 0 && (
          <p className={styles.empty}>Нет попыток по выбранному фильтру</p>
        )}
      </div>
    </div>
  );
};

function renderResult(a: AnalyticsRecentAttempt, mode: ResultMode): string {
  if (a.status !== "completed" || a.percent === undefined) return "—";
  if (mode === "percent") return `${a.percent.toFixed(1)}%`;
  if (a.score === undefined || a.maxScore === undefined) return "—";
  return `${a.score} / ${a.maxScore}`;
}

function sortValue(
  a: AnalyticsRecentAttempt,
  field: SortField,
): string | number {
  switch (field) {
    case "studentName":
      return a.studentName.toLowerCase();
    case "testTitle":
      return a.testTitle.toLowerCase();
    case "status":
      return a.status;
    case "percent":
      return a.percent ?? -1;
    case "completedAt":
      return a.completedAt ? new Date(a.completedAt).getTime() : 0;
  }
}

function percentClass(p?: number): string {
  if (p === undefined) return styles.percentEmpty;
  if (p >= 80) return styles.percentGood;
  if (p < 50) return styles.percentBad;
  return "";
}