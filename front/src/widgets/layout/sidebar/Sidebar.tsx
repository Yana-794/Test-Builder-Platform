import { useEffect, type FC } from "react";
import { NavLink } from "react-router";
import {
  BarChart3,
  FileText,
  LayoutDashboard,
  UserCircle,
  Users,
  X,
} from "lucide-react";
import clsx from "clsx";
import styles from "./Sidebar.module.css";

type SidebarProps = {
  isOpen: boolean;
  onClose: () => void;
};

type NavItem = {
  to: string;
  label: string;
  end?: boolean;
  icon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
};

const NAV_ITEMS: NavItem[] = [
  { to: "/app", label: "Главная", end: true, icon: LayoutDashboard },
  { to: "/app/tests", label: "Мои тесты", icon: FileText },
  { to: "/app/students", label: "Мои ученики", icon: Users },
  { to: "/app/results", label: "Результаты", icon: BarChart3 },
  { to: "/app/profile", label: "Профиль", icon: UserCircle },
];

export const Sidebar: FC<SidebarProps> = ({ isOpen, onClose }) => {
  useEffect(() => {
    if (!isOpen) return;

    const handleKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };

    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", handleKey);

    return () => {
      document.body.style.overflow = prevOverflow;
      document.removeEventListener("keydown", handleKey);
    };
  }, [isOpen, onClose]);

  return (
    <>
      <div
        className={clsx(styles.overlay, isOpen && styles.overlayOpen)}
        onClick={onClose}
        aria-hidden
      />

      <aside
        className={clsx(styles.sidebar, isOpen && styles.sidebarOpen)}
        role="dialog"
        aria-modal="true"
        aria-label="Навигация"
        aria-hidden={!isOpen}
      >
        <div className={styles.header}>
          <span className={styles.title}>Меню</span>
          <button
            type="button"
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Закрыть меню"
          >
            <X size={20} strokeWidth={2} />
          </button>
        </div>

        <nav className={styles.nav}>
          {NAV_ITEMS.map(({ to, label, end, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={end}
              onClick={onClose}
              className={({ isActive }) =>
                clsx(styles.link, isActive && styles.linkActive)
              }
            >
              <Icon size={18} strokeWidth={2} />
              <span>{label}</span>
            </NavLink>
          ))}
        </nav>
      </aside>
    </>
  );
};