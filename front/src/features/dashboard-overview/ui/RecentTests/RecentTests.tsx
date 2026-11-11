import type { FC } from "react";
import { useNavigate } from "react-router";
import { ChevronRight, FileText, Plus } from "lucide-react";
import { formatDate } from "../../../../shared/lib/formatDate";
import { Skeleton } from "../../../../shared/ui/skeleton/Skeleton";
import { StatusBadge } from "../../../../shared/ui/status-badge/StatusBadge";
import styles from "./RecentTests.module.css";
import type { RecentTest } from "../../../../entities/dashboard";

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
        <div className={styles.titleWrap}>
          <span className={styles.titleMark} aria-hidden="true" />
          <h2 className={styles.title}>Последние тесты</h2>
        </div>
        <button
          type="button"
          className={styles.link}
          onClick={() => navigate("/app/tests")}
        >
          Все тесты
          <ChevronRight size={14} strokeWidth={2.4} className={styles.linkIcon} />
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
          <span className={styles.emptyIcon} aria-hidden="true">
            <FileText size={28} strokeWidth={2} />
          </span>
          <p className={styles.emptyTitle}>У вас пока нет тестов</p>
          <p className={styles.emptyText}>
            Создайте первый тест, чтобы начать работу.
          </p>
          <button type="button" className={styles.primary} onClick={onCreate}>
            <Plus size={16} strokeWidth={2.4} />
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
                <span className={styles.rowIcon} aria-hidden="true">
                  <FileText size={16} strokeWidth={2.2} />
                </span>
                <span className={styles.name}>{t.title}</span>
                <span className={styles.badge}>
                  <StatusBadge status={t.status} />
                </span>
                <span className={styles.date}>{formatDate(t.createdAt)}</span>
                <ChevronRight
                  size={16}
                  strokeWidth={2.2}
                  className={styles.chevron}
                  aria-hidden="true"
                />
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
};