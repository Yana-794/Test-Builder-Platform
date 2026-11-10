import { useNavigate } from "react-router";
import emptyImage from "../../assets/empty.png";
import type { FC } from "react";
import { Link, useSearchParams } from "react-router";
import type { TestFilter } from "../../entities/test/model/types";
import styles from "./TestsPage.module.css";
import { useTests } from "../../features/tests/model/hooks/useTests";
import { Plus } from "lucide-react";
import { EmptyState } from "../../shared/ui/empty-state/Empty";

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
        <h1 className={styles.title}>Мои тесты</h1>
        <Link to="/app/tests/new" className={styles.createButton}>
          + Создать тест
        </Link>
      </header>

      <div className={styles.tabs}>
        {FILTERS.map((f) => (
          <button
            key={f.value}
            className={[styles.tab, filter === f.value ? styles.tabActive : ""]
              .join(" ")
              .trim()}
            onClick={() => setParams({ filter: f.value })}
          >
            {f.label}
          </button>
        ))}
      </div>

      {isLoading && <div className={styles.loading}>Загрузка...</div>}

      {isError && (
        <div className={styles.error}>
          Не удалось загрузить тесты.{" "}
          <button onClick={() => refetch()} className={styles.retryButton}>
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
                    <span>{new Date(t.updatedAt).toLocaleString("ru-RU")}</span>
                  </span>
                </div>
                <span className={styles.arrow}>→</span>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};