import type { FC } from "react";
import styles from "./ConflictModal.module.css";

type ConflictModalProps = {
  onRefresh: () => void;
  onStay: () => void;
};

export const ConflictModal: FC<ConflictModalProps> = ({
  onRefresh,
  onStay,
}) => (
  <div className={styles.overlay}>
    <div className={styles.modal}>
      <h2 className={styles.title}>Черновик изменён</h2>
      <p className={styles.message}>
        Черновик был изменён в другой вкладке. Обновите данные, чтобы продолжить
        редактирование.
      </p>
      <div className={styles.actions}>
        <button className={styles.secondary} onClick={onStay}>
          Остаться
        </button>
        <button className={styles.primary} onClick={onRefresh}>
          Обновить
        </button>
      </div>
    </div>
  </div>
);
