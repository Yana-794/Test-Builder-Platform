import type { CSSProperties } from "react";
import styles from "./SectionWordmark.module.css";

interface Props {
  text: string;
  lines?: number;
  intensity?: number;
  size?: string;
  fade?: boolean;
}

export function SectionWordmark({
  text,
  lines = 5,
  intensity = 0.02,
  size = "26vw",
  fade = true,
}: Props) {
  return (
    <div
      className={`${styles.wrap} ${fade ? "" : styles.noFade}`}
      aria-hidden
    >
      <div className={styles.stack}>
        {Array.from({ length: lines }, (_, i) => (
          <div
            key={i}
            className={styles.line}
            style={
              {
                "--i": i,
                "--size": size,
              } as CSSProperties
            }
          >
            {text}
          </div>
        ))}
      </div>

      <div
        className={styles.fade}
        style={{ "--intensity": intensity } as CSSProperties}
      />
    </div>
  );
}