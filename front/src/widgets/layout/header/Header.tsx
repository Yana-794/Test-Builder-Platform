import type { FC } from "react";
import { Link } from "react-router";
import { LogOut, Menu, Plus, User } from "lucide-react";
import { Logo } from "../../../shared";
import styles from "./Header.module.css";

type HeaderProps = {
  userName: string;
  onCreateTest: () => void;
  onOpenSidebar: () => void;
  onLogout: () => void;
};


export const Header: FC<HeaderProps> = ({
  userName,
  onCreateTest,
  onOpenSidebar,
  onLogout,
}) => (
  <header className={styles.header}>
    <div className={styles.left}>
      <button
        type="button"
        className={styles.iconButton}
        onClick={onOpenSidebar}
        aria-label="Открыть меню"
      >
        <Menu size={20} strokeWidth={2} />
      </button>

      <Link to="/app" className={styles.logo} aria-label="На главную">
        <Logo variant="default" />
      </Link>
    </div>

    <div className={styles.right}>
      <button
        type="button"
        className={styles.primaryButton}
        onClick={onCreateTest}
      >
        <Plus size={16} strokeWidth={2.5} />
        <span>Создать тест</span>
      </button>

      <div className={styles.user} title={userName}>
        <span className={styles.avatar} aria-hidden>
         <User  size={20} strokeWidth={2} />
        </span>
        <span className={styles.userName}>{userName}</span>
      </div>

      <button
        type="button"
        className={styles.logoutButton}
        onClick={onLogout}
        aria-label="Выйти"
      >
        <LogOut size={16} strokeWidth={2} />
        <span className={styles.logoutLabel}>Выйти</span>
      </button>
    </div>
  </header>
);