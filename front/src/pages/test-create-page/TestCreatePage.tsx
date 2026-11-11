import type { FC } from "react";
import { FilePlus2 } from "lucide-react";
import { CreateTestForm } from "../../features/tests/ui/CreateTestForm/CreateTestForm";
import styles from "./TestCreatePage.module.css";

export const TestCreatePage: FC = () => {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <span className={styles.heroIcon} aria-hidden="true">
          <FilePlus2 size={22} strokeWidth={2.2} />
        </span>
        <div className={styles.headerContent}>
          <h1 className={styles.title}>Создание теста</h1>
          <p className={styles.subtitle}>
            Заполните основные данные. Вопросы можно добавить после создания
            черновика.
          </p>
        </div>
      </header>

      <CreateTestForm />
    </div>
  );
};