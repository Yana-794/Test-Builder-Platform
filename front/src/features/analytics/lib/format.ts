export function formatRelative(iso?: string): string {
  if (!iso) return "—";
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);

  if (diff < 60) return "только что";
  if (diff < 3600) {
    const m = Math.floor(diff / 60);
    return `${m} ${plural(m, "минуту", "минуты", "минут")} назад`;
  }
  if (diff < 86400) {
    const h = Math.floor(diff / 3600);
    return `${h} ${plural(h, "час", "часа", "часов")} назад`;
  }

  const d = new Date(iso);
  const today = new Date();
  if (d.toDateString() === today.toDateString()) return "сегодня";

  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  if (d.toDateString() === yesterday.toDateString()) return "вчера";

  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}

export function plural(
  n: number,
  one: string,
  few: string,
  many: string,
): string {
  const m10 = n % 10;
  const m100 = n % 100;
  if (m10 === 1 && m100 !== 11) return one;
  if (m10 >= 2 && m10 <= 4 && (m100 < 10 || m100 >= 20)) return few;
  return many;
}

export function formatDuration(seconds?: number): string {
  if (seconds === undefined) return "—";
  if (seconds < 60) return `${seconds}с`;
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  if (m < 60) return s > 0 ? `${m}м ${s}с` : `${m}м`;
  const h = Math.floor(m / 60);
  return `${h}ч ${m % 60}м`;
}

export function computeSummary(
  attempts: { status: string; percent?: number }[],
) {
  const completed = attempts.filter((a) => a.status === "completed");
  const percentages = completed
    .map((a) => a.percent)
    .filter((p): p is number => p !== undefined);

  const avg = percentages.length
    ? percentages.reduce((s, p) => s + p, 0) / percentages.length
    : 0;

  return {
    total: attempts.length,
    completed: completed.length,
    averagePercent: avg,
    bestPercent: percentages.length ? Math.max(...percentages) : 0,
    worstPercent: percentages.length ? Math.min(...percentages) : 0,
    completedRate: attempts.length
      ? (completed.length / attempts.length) * 100
      : 0,
  };
}
