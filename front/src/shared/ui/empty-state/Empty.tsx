import type { FC, ReactNode } from "react";
import clsx from "clsx";
import styles from "./EmptyState.module.css";

type EmptyStateSize = "md" | "lg";

type EmptyStateProps = {
  image?: ReactNode;
  title?: ReactNode;
  description?: ReactNode;
  actions?: ReactNode;
  size?: EmptyStateSize;
  className?: string;
};

export const EmptyState: FC<EmptyStateProps> = ({
  image,
  title,
  description,
  actions,
  size = "md",
  className,
}) => {
  const hasText = Boolean(title || description);

  return (
    <section className={clsx(styles.root, styles[size], className)}>
      {image && <div className={styles.illustration}>{image}</div>}

      {hasText && (
        <div className={styles.text}>
          {title && <h2 className={styles.title}>{title}</h2>}
          {description && <p className={styles.description}>{description}</p>}
        </div>
      )}

      {actions && <div className={styles.actions}>{actions}</div>}
    </section>
  );
};