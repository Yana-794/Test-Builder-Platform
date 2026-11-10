import type { FC } from "react";
import { useNavigate } from "react-router";
import type { DashboardStats } from "../../../../entities/test/model/types";
import { Skeleton } from "../../../../shared/ui/skeleton/Skeleton";
import styles from "./StatsCards.module.css";

type Props = {
  stats?: DashboardStats;
  isLoading: boolean;
  onCardClick?: (filter: "all" | "draft" | "published") => void;
};

const CARDS = [
  { key: "totalTests", label: "Все тесты", filter: "all" },
  { key: "publishedTests", label: "Опубликованные", filter: "published" },
  { key: "draftTests", label: "Черновики", filter: "draft" },
] as const;

export const StatsCards: FC<Props> = ({ stats, isLoading, onCardClick }) => {
  const navigate = useNavigate();

  const handleClick = (filter: "all" | "draft" | "published") => {
    if (onCardClick) {
      onCardClick(filter);
      return;
    }
    navigate(`/app/tests?filter=${filter}`);
  };

  return (
    <section className={styles.grid} aria-label="Показатели">
      {CARDS.map((c) => (
        <button
          key={c.key}
          type="button"
          className={styles.card}
          onClick={() => handleClick(c.filter)}
        >
          <span className={styles.label}>{c.label}</span>
          {isLoading || !stats ? (
            <Skeleton height={36} width="60%" radius={8} />
          ) : (
            <span className={styles.value}>{stats[c.key]}</span>
          )}
        </button>
      ))}
    </section>
  );
};