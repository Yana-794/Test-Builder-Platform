import type { FC } from "react";
import { NavLink } from "react-router";
import {
  BarChart3, FileText, LayoutDashboard,
} from "lucide-react";
import clsx from "clsx";
import styles from "./Sidebar.module.css";

type NavItem = {
  to: string;
  label: string;
  end?: boolean;
  icon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
};

const NAV_ITEMS: NavItem[] = [
  { to: "/app", label: "Главная", end: true, icon: LayoutDashboard },
  { to: "/app/tests", label: "Мои тесты", icon: FileText },
  { to: "/app/results", label: "Результаты", icon: BarChart3 },
];

export const Sidebar: FC = () => (
  <aside className={styles.sidebar} aria-label="Навигация">
    <div className={styles.pillWrap}>
      <span className={styles.shoulderTop} aria-hidden />
      <nav className={styles.nav}>
        {NAV_ITEMS.map(({ to, label, end, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            title={label}
            className={({ isActive }) =>
              clsx(styles.link, isActive && styles.linkActive)
            }
          >
            <span className={styles.icon}>
              <Icon size={20} strokeWidth={2} />
            </span>
            <span className={styles.label}>{label}</span>
          </NavLink>
        ))}
      </nav>
      <span className={styles.shoulderBottom} aria-hidden />
    </div>
  </aside>
);