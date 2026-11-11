import type { FC } from "react";
import {
  CheckCircle2,
  ClipboardList,
  Target,
  TrendingDown,
  Trophy,
  type LucideIcon,
} from "lucide-react";
import { computeSummary } from "../../../../features/analytics/lib/format";
import styles from "./SummaryCards.module.css";
import type { AnalyticsRecentAttempt } from "../../../../entities/analytics";

type Tone = "neutral" | "info" | "primary" | "success" | "danger";

type CardProps = {
  icon: LucideIcon;
  label: string;
  value: string | number;
  hint?: string;
  tone: Tone;
};

const Card: FC<CardProps> = ({ icon: Icon, label, value, hint, tone }) => (
  <div className={`${styles.card} ${styles[tone]}`} role="group">
    <span className={styles.iconChip} aria-hidden="true">
      <Icon size={18} strokeWidth={2.2} />
    </span>
    <div className={styles.body}>
      <span className={styles.value}>{value}</span>
      <span className={styles.label}>{label}</span>
      {hint && <span className={styles.hint}>{hint}</span>}
    </div>
  </div>
);

type Props = { attempts: AnalyticsRecentAttempt[] };

export const SummaryCards: FC<Props> = ({ attempts }) => {
  const s = computeSummary(attempts);

  return (
    <div
      className={styles.grid}
      role="group"
      aria-label="Сводные показатели"
    >
      <Card
        icon={ClipboardList}
        label="Всего попыток"
        value={s.total}
        tone="neutral"
      />
      <Card
        icon={CheckCircle2}
        label="Завершено"
        value={s.completed}
        hint={`${s.completedRate.toFixed(0)}% от всех`}
        tone="info"
      />
      <Card
        icon={Target}
        label="Средний результат"
        value={`${s.averagePercent.toFixed(1)}%`}
        tone="primary"
      />
      <Card
        icon={Trophy}
        label="Лучший"
        value={`${s.bestPercent.toFixed(0)}%`}
        tone="success"
      />
      <Card
        icon={TrendingDown}
        label="Худший"
        value={`${s.worstPercent.toFixed(0)}%`}
        tone="danger"
      />
    </div>
  );
};