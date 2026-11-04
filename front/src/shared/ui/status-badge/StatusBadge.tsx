import type { FC } from "react";
import styles from "./StatusBadge.module.css";
import type { TestStatus } from "../../../entities/test/model/types";

const LABELS: Record<TestStatus, string> = {
  draft: "Черновик",
  published: "Опубликован",
};

interface Props {
  status: TestStatus;
}

export const StatusBadge: FC<Props> = ({ status }) => (
  <span className={`${styles.badge} ${styles[status]}`}>{LABELS[status]}</span>
);