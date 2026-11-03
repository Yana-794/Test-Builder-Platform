import type { FC } from "react";
import { useNavigate } from "react-router";
import { Plus, Users } from "lucide-react";
import { EmptyState } from "../../shared/ui/empty-state/Empty"
import { useAuth } from "../../features/auth/hooks/useAuth";
import emptyImage from "../../assets/Empty.png";
import styles from "./DashboardPage.module.css";

export const DashboardPage: FC = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const userName = user?.display_name ?? "пользователь";

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Приветствую, {userName}</h1>
        <p className={styles.subtitle}>
          Всё для вашего тестирования в одном месте.
        </p>
      </header>

      <EmptyState
        size="lg"
        image={<img src={emptyImage} alt="" />}
        actions={
          <>
            <button
              type="button"
              className={styles.primaryButton}
              onClick={() => navigate("/app/tests/new")}
            >
              <Plus size={16} strokeWidth={2.5} />
              Создать тест
            </button>
            <button
              type="button"
              className={styles.secondaryButton}
              onClick={() => navigate("/app/students")}
            >
              <Users size={16} strokeWidth={2} />
              Создать класс
            </button>
          </>
        }
      />
    </div>
  );
};