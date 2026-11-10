import { useEffect, useRef } from "react";
import styles from "./TestPreview.module.css";

export function TestPreviewSection() {
  const innerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = innerRef.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.classList.add(styles.innerVisible);
          observer.disconnect();
        }
      },
      { threshold: 0.2 },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <section className={styles.section}>


      <div ref={innerRef} className={styles.inner}>
        <div className={styles.text}>
          <span className={styles.eyebrow}>
            <span className={styles.eyebrowDot} aria-hidden />
            Для учителя и ученика
          </span>

          <h2 className={styles.title}>
           <span className={styles.titleAccent}>Понятно</span> вам.  <span className={styles.titleAccent}>Удобно</span> {" "}
           вашим ученикам.
          </h2>

          <p className={styles.subtitle}>
            Создавайте тесты в удобном редакторе, а ученики будут проходить их в
            простом интерфейсе без лишних шагов.
          </p>
        </div>

        <div className={styles.preview}>
          <div className={styles.previewGlow} aria-hidden />

          <div className={styles.previewCard}>
            <div className={styles.previewHeader}>
              <span className={styles.previewTitle}>Основы JavaScript</span>
              <span className={styles.previewCounter}>Вопрос 1 из 5</span>
            </div>

            <div className={styles.progress}>
              <div className={styles.progressFill} style={{ width: "20%" }} />
            </div>

            <p className={styles.question}>
              Как объявить переменную в JavaScript?
            </p>

            <div className={styles.options}>
              <label className={styles.option}>
                <span className={styles.optionRadio} />
                <span className={styles.optionLabel}>var</span>
              </label>

              <label className={`${styles.option} ${styles.optionSelected}`}>
                <span className={styles.optionRadio} />
                <span className={styles.optionLabel}>const</span>
              </label>

              <label className={styles.option}>
                <span className={styles.optionRadio} />
                <span className={styles.optionLabel}>let</span>
              </label>
            </div>

            <div className={styles.previewFooter}>
              <button className={styles.previewButton} type="button">
                Далее →
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
