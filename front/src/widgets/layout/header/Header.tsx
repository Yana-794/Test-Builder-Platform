import type { FC } from "react";
import { LogOut, Plus, User } from "lucide-react";
import { Logo } from "../../../shared";
import styles from "./Header.module.css";

type HeaderProps = {
  userName: string;
  onCreateTest: () => void;
  onLogout: () => void;
};

export const Header: FC<HeaderProps> = ({
  userName,
  onCreateTest,
  onLogout,
}) => (
  <header className={styles.header}>
    <div className={styles.left}>
      <Logo variant="default" to="/app" />
    </div>

    <div className={styles.right}>
      <button
        type="button"
        className={styles.primaryButton}
        onClick={onCreateTest}
        title="Создать тест"
        aria-label="Создать тест"
      >
        <Plus size={16} strokeWidth={2.5} aria-hidden="true" />
        <span className={styles.primaryLabel}>Создать тест</span>
      </button>

      <div className={styles.user} title={userName}>
        <span className={styles.avatar} aria-hidden="true">
          <User size={18} strokeWidth={2.2} />
        </span>
        <span className={styles.userName}>{userName}</span>
      </div>

      <button
        type="button"
        className={styles.logoutButton}
        onClick={onLogout}
        aria-label="Выйти"
        title="Выйти"
      >
        <LogOut size={16} strokeWidth={2.2} aria-hidden="true" />
        <span className={styles.logoutLabel}>Выйти</span>
      </button>
    </div>
  </header>
);