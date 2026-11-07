import type { FC } from "react";
import { CreateTestForm } from "../../features/tests/ui/CreateTestForm/CreateTestForm";
import styles from "./TestCreatePage.module.css";

export const TestCreatePage: FC = () => {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Создание теста</h1>
        <p className={styles.subtitle}>
          Заполните основные данные. Вопросы можно добавить после создания
          черновика.
        </p>
      </header>

      <CreateTestForm />
    </div>
  );
};
