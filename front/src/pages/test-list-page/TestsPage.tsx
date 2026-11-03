import type { FC } from "react";
import { useNavigate } from "react-router";
import { Plus } from "lucide-react";
import { EmptyState } from "../../shared/ui/empty-state/Empty";
import emptyImage from "../../assets/Empty.png";
import styles from "./TestsPage.module.css";

export const TestsPage: FC = () => {
  const navigate = useNavigate();
  const createTest = () => navigate("/app/tests/new");

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.headerText}>
          <h1 className={styles.title}>Мои тесты</h1>
          <p className={styles.subtitle}>
            Здесь будут все ваши тесты: черновики и опубликованные.
          </p>
        </div>

        <button
          type="button"
          className={styles.primaryButton}
          onClick={createTest}
        >
          <Plus size={16} strokeWidth={2.5} />
          Создать тест
        </button>
      </header>

      <EmptyState
        image={<img src={emptyImage} alt="" />}
        title="Пока нет тестов"
        description="Создайте первый тест, чтобы поделиться им с учениками и увидеть результаты."
        actions={
          <button
            type="button"
            className={styles.primaryButton}
            onClick={createTest}
          >
            <Plus size={16} strokeWidth={2.5} />
            Создать тест
          </button>
        }
      />
    </div>
  );
};