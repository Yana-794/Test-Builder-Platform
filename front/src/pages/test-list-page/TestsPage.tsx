import type { FC } from "react";
import { Link, useNavigate, useSearchParams } from "react-router";
import { AlertCircle, ChevronRight, FileText, Plus } from "lucide-react";
import emptyImage from "../../assets/empty.png";
import type { TestFilter } from "../../entities/test/model/types";
import { useTests } from "../../features/tests/model/hooks/useTests";
import { EmptyState } from "../../shared/ui/empty-state/Empty";
import { Skeleton } from "../../shared/ui/skeleton/Skeleton";
import styles from "./TestsPage.module.css";

const FILTERS: { value: TestFilter; label: string }[] = [
  { value: "all", label: "Все" },
  { value: "draft", label: "Черновики" },
  { value: "published", label: "Опубликованные" },
];

const STATUS_LABEL: Record<string, string> = {
  draft: "Черновик",
  published: "Опубликован",
  archived: "Архив",
};

export const TestsPage: FC = () => {
  const [params, setParams] = useSearchParams();
  const raw = params.get("filter") ?? "all";
  const filter: TestFilter = (
    ["all", "draft", "published", "archived"].includes(raw) ? raw : "all"
  ) as TestFilter;

  const { data, isLoading, isError, refetch } = useTests(filter);

  const navigate = useNavigate();
  const createTest = () => navigate("/app/tests/new");

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.headerContent}>
          <h1 className={styles.title}>Мои тесты</h1>
          <p className={styles.subtitle}>
            Управляйте тестами, черновиками и публикациями в одном месте.
          </p>
        </div>
        <Link to="/app/tests/new" className={styles.createButton}>
          <Plus size={16} strokeWidth={2.5} />
          Создать тест
        </Link>
      </header>

      <div className={styles.tabs} role="tablist" aria-label="Фильтр тестов">
        {FILTERS.map((f) => {
          const active = filter === f.value;
          return (
            <button
              key={f.value}
              type="button"
              role="tab"
              aria-selected={active}
              className={`${styles.tab} ${active ? styles.tabActive : ""}`}
              onClick={() => setParams({ filter: f.value })}
            >
              {f.label}
            </button>
          );
        })}
      </div>

      {isLoading && (
        <ul className={styles.list} aria-busy="true" aria-live="polite">
          {Array.from({ length: 4 }).map((_, i) => (
            <li key={i}>
              <Skeleton height={68} radius={14} />
            </li>
          ))}
        </ul>
      )}

      {isError && (
        <div className={styles.error} role="alert">
          <span className={styles.errorIcon} aria-hidden="true">
            <AlertCircle size={22} strokeWidth={2.2} />
          </span>
          <div className={styles.errorBody}>
            <p className={styles.errorTitle}>Не удалось загрузить тесты</p>
            <p className={styles.errorText}>
              Проверьте соединение и попробуйте ещё раз.
            </p>
          </div>
          <button
            type="button"
            onClick={() => refetch()}
            className={styles.retryButton}
          >
            Повторить
          </button>
        </div>
      )}

      {!isLoading && !isError && data && data.tests.length === 0 && (
        <EmptyState
          image={<img src={emptyImage} alt="" />}
          title="Пока нет тестов"
          description="Создайте первый тест, чтобы поделиться им с учениками и увидеть результаты."
          actions={
            <button
              type="button"
              className={styles.primaryButton}
              onClick={createTest}
            >
              <Plus size={16} strokeWidth={2.5} />
              Создать тест
            </button>
          }
        />
      )}

      {!isLoading && !isError && data && data.tests.length > 0 && (
        <ul className={styles.list}>
          {data.tests.map((t) => (
            <li key={t.id}>
              <Link to={`/app/tests/${t.id}`} className={styles.item}>
                <span className={styles.itemIcon} aria-hidden="true">
                  <FileText size={18} strokeWidth={2.2} />
                </span>

                <div className={styles.itemMain}>
                  <span className={styles.itemTitle}>
                    {t.title || "(без названия)"}
                  </span>
                  <span className={styles.itemMeta}>
                    <span
                      className={`${styles.badge} ${styles[`badge_${t.status}`]}`}
                    >
                      {STATUS_LABEL[t.status] ?? t.status}
                    </span>
                    <span className={styles.revision}>v{t.revision}</span>
                    <span className={styles.dot} aria-hidden="true" />
                    <span className={styles.date}>
                      {new Date(t.updatedAt).toLocaleString("ru-RU")}
                    </span>
                  </span>
                </div>

                <ChevronRight
                  size={18}
                  strokeWidth={2.2}
                  className={styles.chevron}
                  aria-hidden="true"
                />
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};