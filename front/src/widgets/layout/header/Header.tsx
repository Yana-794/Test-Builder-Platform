import type { FC } from "react";
import { Plus } from "lucide-react";
import { Logo } from "../../../shared";
import styles from "./Header.module.css";
import { UserMenu } from "./UserMenu";
import type { User } from "../../../features/auth/api/types";
import { Link } from "react-router";

type HeaderProps = {
  user: User | null;
  onCreateTest: () => void;
  onLogout: () => void;
};

export const Header: FC<HeaderProps> = ({ user, onCreateTest, onLogout }) => (
  <header className={styles.header}>
    <div className={styles.left}>
      <Link to="/app" className={styles.logo} aria-label="На главную">
        <Logo variant="default" />
      </Link>
    </div>

    <div className={styles.right}>
      <button className={styles.primaryButton} onClick={onCreateTest}>
        <Plus size={16} strokeWidth={2.5} />
        <span>Создать тест</span>
      </button>

      {user && <UserMenu user={user} onLogout={onLogout} />}
    </div>
  </header>
);
