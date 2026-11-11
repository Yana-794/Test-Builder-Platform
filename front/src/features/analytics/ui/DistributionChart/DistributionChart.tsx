import type { FC } from "react";
import { useMemo } from "react";
import { BarChart3 } from "lucide-react";
import styles from "./DistributionChart.module.css";
import type { AnalyticsRecentAttempt } from "../../../../entities/analytics";

type Bucket = {
  label: string;
  min: number;
  max: number;
  color: string;
  gradient: string;
};


const BUCKETS: Bucket[] = [
  { label: "80–100%", min: 80, max: 100.01, color: "#10b981", gradient: "linear-gradient(90deg, #10b981 0%, #34d399 100%)" },
  { label: "60–80%",  min: 60, max: 80.01,  color: "#84cc16", gradient: "linear-gradient(90deg, #84cc16 0%, #a3e635 100%)" },
  { label: "40–60%",  min: 40, max: 60.01,  color: "#eab308", gradient: "linear-gradient(90deg, #eab308 0%, #facc15 100%)" },
  { label: "20–40%",  min: 20, max: 40.01,  color: "#f59e0b", gradient: "linear-gradient(90deg, #f59e0b 0%, #fbbf24 100%)" },
  { label: "0–20%",   min: 0,  max: 20.01,  color: "#ef4444", gradient: "linear-gradient(90deg, #ef4444 0%, #f87171 100%)" },
];

type Props = { attempts: AnalyticsRecentAttempt[] };

export const DistributionChart: FC<Props> = ({ attempts }) => {
  const { rows, total, maxCount } = useMemo(() => {
    const completed = attempts.filter(
      (a) => a.status === "completed" && a.percent !== undefined,
    );

    const buckets = BUCKETS.map((b) => ({
      ...b,
      count: completed.filter((a) => a.percent! >= b.min && a.percent! < b.max).length,
    }));

    const total = buckets.reduce((s, b) => s + b.count, 0);
    const maxCount = Math.max(...buckets.map((b) => b.count), 1);

    return { rows: buckets, total, maxCount };
  }, [attempts]);

  if (total === 0) {
    return (
      <div className={styles.empty}>
        <BarChart3 size={32} strokeWidth={1.5} className={styles.emptyIcon} />
        <p className={styles.emptyText}>Нет завершённых попыток для графика</p>
      </div>
    );
  }

  return (
    <div className={styles.wrap}>

      <div className={styles.rows}>
        {rows.map((r, i) => {
          const percent = (r.count / total) * 100;
          const width = (r.count / maxCount) * 100;
          const isEmpty = r.count === 0;

          return (
            <div
              key={r.label}
              className={[styles.row, isEmpty && styles.rowEmpty].join(" ").trim()}
            >
              <span className={styles.label}>
                <span className={styles.dot} style={{ background: r.color }} />
                {r.label}
              </span>

              <div className={styles.track}>
                <div
                  className={styles.fill}
                  style={{
                    ["--bar-width" as string]: `${width}%`,
                    ["--delay" as string]: `${i * 60}ms`,
                    background: r.gradient,
                  }}
                />
              </div>

              <span className={styles.count}>
                <span className={styles.countValue}>{r.count}</span>
                <span className={styles.countPercent}>{percent.toFixed(0)}%</span>
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
};