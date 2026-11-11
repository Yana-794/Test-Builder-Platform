import type { FC } from "react";
import { useNavigate } from "react-router";
import { LayoutGrid, CheckCircle2, FileEdit, ArrowRight } from "lucide-react";
import { Skeleton } from "../../../../shared/ui/skeleton/Skeleton";
import styles from "./StatsCards.module.css";
import type { DashboardStats } from "../../../../entities/dashboard";

type Props = {
  stats?: DashboardStats;
  isLoading: boolean;
  onCardClick?: (filter: "all" | "draft" | "published") => void;
};

const CARDS = [
  {
    key: "totalTests",
    label: "Все тесты",
    filter: "all",
    icon: LayoutGrid,
    tone: "primary",
  },
  {
    key: "publishedTests",
    label: "Опубликованные",
    filter: "published",
    icon: CheckCircle2,
    tone: "success",
  },
  {
    key: "draftTests",
    label: "Черновики",
    filter: "draft",
    icon: FileEdit,
    tone: "warning",
  },
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
      {CARDS.map((c) => {
        const Icon = c.icon;
        return (
          <button
            key={c.key}
            type="button"
            className={`${styles.card} ${styles[c.tone]}`}
            onClick={() => handleClick(c.filter)}
          >
            <span className={styles.top}>
              <span className={styles.iconWrap} aria-hidden="true">
                <Icon size={18} strokeWidth={2.2} />
              </span>
              <span className={styles.label}>{c.label}</span>
            </span>

            {isLoading || !stats ? (
              <Skeleton height={36} width="60%" radius={8} />
            ) : (
              <span className={styles.value}>{stats[c.key]}</span>
            )}

            <span className={styles.hint} aria-hidden="true">
              Смотреть <ArrowRight size={14} strokeWidth={2.4} />
            </span>
          </button>
        );
      })}
    </section>
  );
};