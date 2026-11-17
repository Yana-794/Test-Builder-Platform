import type { FC } from "react";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router";
import { LogOut, UserCircle } from "lucide-react";
import type { User } from "../../../features/auth/api/types";
import styles from "./UserMenu.module.css";

type Props = {
  user: User;
  onLogout: () => void;
};

export const UserMenu: FC<Props> = ({ user, onLogout }) => {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const hoverTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const openMenu = () => {
    if (hoverTimer.current) clearTimeout(hoverTimer.current);
    setOpen(true);
  };

  const closeMenuDelayed = () => {
    hoverTimer.current = setTimeout(() => setOpen(false), 150);
  };

  // click outside
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  // Esc
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => { if (e.key === "Escape") setOpen(false); };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open]);

  const initial = (user.display_name || user.email).charAt(0).toUpperCase();

  return (
    <div
      className={styles.root}
      ref={rootRef}
      onMouseEnter={openMenu}
      onMouseLeave={closeMenuDelayed}
    >
      <button
        type="button"
        className={styles.trigger}
        onClick={() => setOpen((v) => !v)}
        aria-haspopup="menu"
        aria-expanded={open}
        aria-label="Меню пользователя"
      >
        <span className={styles.avatar}>{initial}</span>
        <span className={styles.name}>{user.display_name}</span>
      </button>

      {open && (
        <div className={styles.menu} role="menu">
          <div className={styles.menuHeader}>
            <span className={styles.menuName}>{user.display_name}</span>
            <span className={styles.menuEmail}>{user.email}</span>
          </div>

          <Link
            to="/app/profile"
            role="menuitem"
            className={styles.menuItem}
            onClick={() => setOpen(false)}
          >
            <UserCircle size={16} strokeWidth={2} />
            Профиль
          </Link>

          <button
            type="button"
            role="menuitem"
            className={styles.menuItem}
            onClick={() => { setOpen(false); onLogout(); }}
          >
            <LogOut size={16} strokeWidth={2} />
            Выйти
          </button>
        </div>
      )}
    </div>
  );
};