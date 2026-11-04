import type { FC } from "react";
import styles from "./Skeleton.module.css";

interface Props {
  height?: number;
  width?: string;
  radius?: number;
}

export const Skeleton: FC<Props> = ({
  height = 16,
  width = "100%",
  radius = 6,
}) => (
  <span
    className={styles.skeleton}
    style={{ height, width, borderRadius: radius }}
    aria-hidden="true"
  />
);