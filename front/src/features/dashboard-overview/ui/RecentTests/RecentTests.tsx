import type { FC } from "react";
import { useNavigate } from "react-router";
import type { RecentTest } from "../../../../entities/test/model/types";
import { formatDate } from "../../../../shared/lib/formatDate";
import { Skeleton } from "../../../../shared/ui/skeleton/Skeleton";
import { StatusBadge } from "../../../../shared/ui/status-badge/StatusBadge";
import styles from "./RecentTests.module.css";

interface Props {
  items?: RecentTest[];
  isLoading: boolean;
  onCreate: () => void;
}

export const RecentTests: FC<Props> = ({ items, isLoading, onCreate }) => {
  const navigate = useNavigate();

  return (
    <section className={styles.section} aria-label="Последние тесты">
      <header className={styles.head}>
        <h2 className={styles.title}>Последние тесты</h2>
        <button
          type="button"
          className={styles.link}
          onClick={() => navigate("/app/tests")}
        >
          Все тесты →
        </button>
      </header>

      {isLoading ? (
        <div className={styles.list}>
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} height={56} radius={12} />
          ))}
        </div>
      ) : !items || items.length === 0 ? (
        <div className={styles.empty}>
          <p className={styles.emptyTitle}>У вас пока нет тестов</p>
          <p className={styles.emptyText}>
            Создайте первый тест, чтобы начать работу.
          </p>
          <button type="button" className={styles.primary} onClick={onCreate}>
            Создать тест
          </button>
        </div>
      ) : (
        <ul className={styles.list}>
          {items.map((t) => (
            <li key={t.id}>
              <button
                type="button"
                className={styles.row}
                onClick={() => navigate(`/app/tests/${t.id}`)}
              >
                <span className={styles.name}>{t.title}</span>
                <StatusBadge status={t.status} />
                <span className={styles.date}>{formatDate(t.createdAt)}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
};
