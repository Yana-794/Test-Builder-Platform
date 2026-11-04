import type { FC } from "react";
import styles from "./ErrorState.module.css";

interface Props {
  onRetry: () => void;
  message?: string;
  retryLabel?: string;
}

export const ErrorState: FC<Props> = ({
  onRetry,
  message = "Не удалось загрузить данные",
  retryLabel = "Повторить",
}) => (
  <div className={styles.wrap} role="alert">
    <p className={styles.text}>{message}</p>
    <button type="button" className={styles.retry} onClick={onRetry}>
      {retryLabel}
    </button>
  </div>
);